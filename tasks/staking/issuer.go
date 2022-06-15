package staking

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/logger"
	"github.com/pkg/errors"
	"time"
)

type Issuer struct {
	logger.Logger
	CChainIssueClient chain.Issuer
	PChainIssueClient chain.Issuer
}

func (i *Issuer) IssueTask(ctx context.Context, task *StakeTask) ([]ids.ID, error) {
	txBytesArr, err := i.getTxBytes(task)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ids, err := i.doIssue(ctx, txBytesArr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return ids, nil
}

func (i *Issuer) doIssue(ctx context.Context, txBytesArr [][]byte) ([]ids.ID, error) {
	exportId, err := i.CChainIssueClient.IssueTx(ctx, txBytesArr[0])
	if err != nil {
		i.Error("Issuer failed to issue signed exportTx", logger.Field{"error", err})
		return nil, errors.Wrap(err, "failed to issue signed exportTx")
	}

	// sleep to avoid error: "failed to get shared memory: not found"
	time.Sleep(time.Second * 5)
	importId, err := i.PChainIssueClient.IssueTx(ctx, txBytesArr[1])
	if err != nil {
		i.Error("Stake failed to issue signed importTx", logger.Field{"error", err})
		return nil, errors.Wrap(err, "failed to issue signed importTx")
	}

	// sleep to avoid error: "failed to get shared memory: not found"
	time.Sleep(time.Second * 5)
	addDelegatorId, err := i.PChainIssueClient.IssueTx(ctx, txBytesArr[2])
	if err != nil {
		i.Error("Stake failed to issue signed addDelegatorTx", logger.Field{"error", err})
		return nil, errors.Wrap(err, "failed to issue signed importTx")
	}

	return []ids.ID{exportId, importId, addDelegatorId}, nil
}

func (i *Issuer) getTxBytes(task *StakeTask) ([][]byte, error) {
	exportTx, err := task.GetSignedExportTx()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get signed exportTx")
	}

	importTx, err := task.GetSignedImportTx()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get signed importTx")
	}

	addDelegatorTx, err := task.GetSignedAddDelegatorTx()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get signed addDelegatorTx")
	}

	return [][]byte{exportTx.Bytes(), importTx.Bytes(), addDelegatorTx.Bytes()}, nil
}
