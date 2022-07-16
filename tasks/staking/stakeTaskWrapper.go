package staking

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/logger"
	"github.com/pkg/errors"
	"time"
)

// todo: consider refactoring with Template Method design pattern

type StakeTaskWrapper struct {
	*SignRequester
	*StakeTask
	CChainIssueClient chain.CChainIssuer
	Logger            logger.Logger
	PChainIssueClient chain.PChainIssuer
}

func (s *StakeTaskWrapper) SignTx(ctx context.Context) error {
	// ExportTx
	txHash, err := s.ExportTxHash()
	if err != nil {
		return errors.WithStack(err)
	}

	s.Logger.Debug("Signing exportTx", []logger.Field{{}}...)

	sig, err := s.SignExportTx(ctx, txHash)
	if err != nil {
		return errors.WithStack(err)
	}

	s.Logger.Debug("Signed exportTx", []logger.Field{{}}...)

	err = s.SetExportTxSig(sig)
	if err != nil {
		return errors.WithStack(err)
	}

	// ImportTx
	txHash, err = s.ImportTxHash()
	if err != nil {
		return errors.WithStack(err)
	}

	s.Logger.Debug("Signing importTx", []logger.Field{{}}...)

	sig, err = s.SignImportTx(ctx, txHash)
	if err != nil {
		return errors.WithStack(err)
	}

	s.Logger.Debug("Signed importTx", []logger.Field{{}}...)

	err = s.SetImportTxSig(sig)
	if err != nil {
		return errors.WithStack(err)
	}

	// AddDelegatorTx
	txHash, err = s.AddDelegatorTxHash()
	if err != nil {
		return errors.WithStack(err)
	}

	s.Logger.Debug("Signing addDelegatorTx", []logger.Field{{}}...)

	sig, err = s.SignAddDelegatorTx(ctx, txHash)
	if err != nil {
		return errors.WithStack(err)
	}

	s.Logger.Debug("Signed addDelegatorTx", []logger.Field{{}}...)

	err = s.SetAddDelegatorTxSig(sig)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *StakeTaskWrapper) IssueTx(ctx context.Context) ([]ids.ID, error) {
	// ExportTx
	exportTx, err := s.GetSignedExportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	exportId, err := s.CChainIssueClient.IssueTx(ctx, exportTx.SignedBytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	s.Logger.Debug("Issued exportTx from C-Chain", []logger.Field{{"exportTxCChain", exportId}}...)

	// ImportTx
	time.Sleep(time.Second * 10) // sleep to avoid error: "failed to get shared memory"

	importTx, err := s.GetSignedImportTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	importId, err := s.PChainIssueClient.IssueTx(ctx, importTx.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	s.Logger.Debug("Issued importTx to P-Chain", []logger.Field{{"importTxPChain", importId}}...)

	// AddDelegatorTx
	time.Sleep(time.Second * 10) // sleep to avoid error: "failed to get shared memory"

	addDelegatorTx, err := s.GetSignedAddDelegatorTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	addDelegatorId, err := s.PChainIssueClient.IssueTx(ctx, addDelegatorTx.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	s.Logger.Debug("Issued addDelegatorTx to P-Chain", []logger.Field{{"addDelegatorTxPChain", addDelegatorId}}...)

	return []ids.ID{exportId, importId, addDelegatorId}, nil
}
