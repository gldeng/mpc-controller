package porter

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/port/issuer"
	"github.com/pkg/errors"
)

const (
	ErrMsgFailedToVerifySig = "failed to verify signature"
)

type Txs interface {
	ExportTx
	ImportTx
}

type ExportTx interface {
	ExportTxHash() ([]byte, error)
	SetExportTxSig(exportTxSig [65]byte) error
	SignedExportTxBytes() ([]byte, error)
}

type ImportTx interface {
	ImportTxHash() ([]byte, error)
	SetImportTxSig(importTxSig [65]byte) error
	SignedImportTxBytes() ([]byte, error)
}

type TxSigner interface {
	SignExportTx(ctx context.Context, exportTxHash []byte) ([65]byte, error)
	SignImportTx(ctx context.Context, importTxHash []byte) ([65]byte, error)
}

type SigVerifier interface {
	VerifySig(hash []byte, signature [65]byte) (bool, error)
}

type TxIssuer interface {
	IssueExportTx(ctx context.Context, exportTxBytes []byte) (ids.ID, issuer.IssueOrder, error)
	IssueImportTx(ctx context.Context, importTxBytes []byte) (ids.ID, issuer.IssueOrder, error)
}

type Porter struct {
	logger.Logger
	Txs
	TxSigner
	TxIssuer
	SigVerifier
}

func (p *Porter) SignAndIssueTxs(ctx context.Context) ([2]ids.ID, error) {
	// Sign ExportTx
	exportTxHash, err := p.ExportTxHash()
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	exportTxSig, err := p.SignExportTx(ctx, exportTxHash)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	ok, err := p.VerifySig(exportTxHash, exportTxSig)
	if err != nil {
		return [2]ids.ID{}, errors.Wrap(err, ErrMsgFailedToVerifySig)
	}

	if !ok {
		return [2]ids.ID{}, errors.Wrapf(err, "failed to verify ExportTx signature, hashHex:%q, sigHex:%q",
			bytes.BytesToHex(exportTxHash), bytes.Bytes65ToHex(exportTxSig))
	}

	err = p.SetExportTxSig(exportTxSig)
	if err != nil {
		return [2]ids.ID{}, errors.Wrapf(err, "failed to set exportTx signature")
	}

	// Sign ImportTx
	importTxHash, err := p.ImportTxHash()
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	importTxSig, err := p.SignImportTx(ctx, importTxHash)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	ok, err = p.VerifySig(importTxHash, importTxSig)
	if err != nil {
		return [2]ids.ID{}, errors.Wrap(err, ErrMsgFailedToVerifySig)
	}

	if !ok {
		return [2]ids.ID{}, errors.Wrapf(err, "failed to verify ImportTx signature, hashHex:%q, sigHex:%q",
			bytes.BytesToHex(importTxHash), bytes.Bytes65ToHex(importTxSig))
	}

	err = p.SetImportTxSig(importTxSig)
	if err != nil {
		return [2]ids.ID{}, errors.Wrapf(err, "failed to set importTx signature")
	}

	//n := rand.Intn(10_000_000_000)
	//time.Sleep(time.Nanosecond * time.Duration(n)) // reduce concurrent conflict

	// Issue ExportTx
	exportTxBytes, err := p.SignedExportTxBytes()
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	exportTxId, _, err := p.IssueExportTx(ctx, exportTxBytes)
	if err != nil {
		return [2]ids.ID{}, errors.Wrapf(err, "failed to IssueExportTx")
	}

	//time.Sleep(time.Second * 5) // wait for shared memory get ready

	// Issue ImportTx
	importTxBytes, err := p.SignedImportTxBytes()
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	importTxId, _, err := p.IssueImportTx(ctx, importTxBytes)
	if err != nil {
		return [2]ids.ID{}, errors.Wrapf(err, "failed to IssueImportTx")
	}

	return [2]ids.ID{exportTxId, importTxId}, nil
}
