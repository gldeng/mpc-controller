package task

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/logger"
	"github.com/pkg/errors"
	"time"
)

type Staker struct {
	CChainClient evm.Client
	PChainClient platformvm.Client
}

func NewStaker(cChainUri, pChainUri string) *Staker {
	if cChainUri == "" {
		cChainUri = "http://localhost:9650"
	}

	if pChainUri == "" {
		pChainUri = "http://localhost:9650"
	}

	// todo: check whether clients can connect normally.
	cclient := evm.NewClient(cChainUri, "C")
	pclient := platformvm.NewClient(pChainUri)

	return &Staker{
		CChainClient: cclient,
		PChainClient: pclient,
	}
}

func (s *Staker) IssueStakeTaskTxs(ctx context.Context, task *StakeTask) ([]ids.ID, error) {
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

	ids, err := s.IssueSignedStakeTxs(ctx, exportTx.Bytes(), importTx.Bytes(), addDelegatorTx.Bytes())
	return ids, errors.Wrap(err, "failed to issue stake task txs")
}

func (s *Staker) IssueSignedStakeTxs(ctx context.Context, exportTx, importTx, addDelegatorTx []byte) ([]ids.ID, error) {
	exportId, err := s.CChainClient.IssueTx(ctx, exportTx)
	if err != nil {
		logger.Error("Staker failed to issue signed exportTx",
			logger.Field{"error", err})
		return nil, errors.Wrap(err, "failed to issue signed exportTx")
	}

	time.Sleep(time.Second * 2)
	importId, err := s.PChainClient.IssueTx(ctx, importTx)
	if err != nil {
		logger.Error("Stake failed to issue signed importTx",
			logger.Field{"error", err})
		return nil, errors.Wrap(err, "failed to issue signed importTx")
	}

	time.Sleep(time.Second * 2)
	addDelegatorId, err := s.PChainClient.IssueTx(ctx, addDelegatorTx)
	if err != nil {
		logger.Error("Stake failed to issue signed addDelegatorTx",
			logger.Field{"error", err})
		return nil, errors.Wrap(err, "failed to issue signed importTx")
	}

	return []ids.ID{exportId, importId, addDelegatorId}, nil
}
