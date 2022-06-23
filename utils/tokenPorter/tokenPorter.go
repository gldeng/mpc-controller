package tokenPorter

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/pkg/errors"
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

type TxIssuer interface {
	IssueExportTx(ctx context.Context, exportTxBytes []byte) (ids.ID, error)
	IssueImportTx(ctx context.Context, importTxBytes []byte) (ids.ID, error)
}

type TokenPorter struct {
	Txs
	TxSigner
	TxIssuer
}

func (p *TokenPorter) SignAndTransferTxs(ctx context.Context) ([2]ids.ID, error) {
	// Sign ExportTx
	exportTxHash, err := p.ExportTxHash()
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	exportTxSig, err := p.SignExportTx(ctx, exportTxHash)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	err = p.SetExportTxSig(exportTxSig)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
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

	err = p.SetImportTxSig(importTxSig)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	// Issue ExportTx
	exportTxBytes, err := p.SignedExportTxBytes()
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	exportTxId, err := p.IssueExportTx(ctx, exportTxBytes)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	// Issue ImportTx
	importTxBytes, err := p.SignedImportTxBytes()
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	importTxId, err := p.IssueImportTx(ctx, importTxBytes)
	if err != nil {
		return [2]ids.ID{}, errors.WithStack(err)
	}

	return [2]ids.ID{exportTxId, importTxId}, nil
}
