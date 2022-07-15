package staking

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
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

	sig, err := s.SignExportTx(ctx, txHash)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.SetExportTxSig(sig)
	if err != nil {
		return errors.WithStack(err)
	}

	// ImportTx
	txHash, err = s.ImportTxHash()
	if err != nil {
		return errors.WithStack(err)
	}

	sig, err = s.SignImportTx(ctx, txHash)
	if err != nil {
		return errors.WithStack(err)
	}

	err = s.SetImportTxSig(sig)
	if err != nil {
		return errors.WithStack(err)
	}

	// AddDelegatorTx
	txHash, err = s.AddDelegatorTxHash()
	if err != nil {
		return errors.WithStack(err)
	}

	sig, err = s.SignAddDelegatorTx(ctx, txHash)
	if err != nil {
		return errors.WithStack(err)
	}

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

	// AddDelegatorTx
	time.Sleep(time.Second * 10) // sleep to avoid error: "failed to get shared memory"

	addDelegatorTx, err := s.GetSignedAddDelegatorTx()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	addDelegatorId, err := s.PChainIssueClient.IssueTx(ctx, addDelegatorTx.Bytes()) // todo: reuse chain.PlatformvmClientWrapper
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = s.awaitAddDelegatorTxDecided(ctx, addDelegatorId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return []ids.ID{exportId, importId, addDelegatorId}, nil
}

func (s *StakeTaskWrapper) awaitAddDelegatorTxDecided(ctx context.Context, txID ids.ID) (err error) { // todo: reuse chain.PlatformvmClientWrapper
	var txStatusRes *platformvm.GetTxStatusResponse
	backoff.RetryFnExponentialForever(s.Logger, ctx, func() error {
		txStatusRes, err = s.PChainIssueClient.(platformvm.Client).AwaitTxDecided(ctx, txID, time.Second)
		if err != nil {
			err = errors.WithStack(err)
			return err
		}
		return nil
	})

	if txStatusRes.Status.String() != "Committed" {
		return errors.Errorf("addDelegatorTx failed. txID:%q, status:%q, reason:%q",
			bytes.Bytes32ToHex(txID), txStatusRes.Status.String(), txStatusRes.Reason)
	}

	return nil
}
