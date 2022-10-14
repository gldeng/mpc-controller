package txissuer

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"time"
)

const (
	StatusCreated Status = iota
	StatusIssued
	StatusProcessing
	StatusApproved
	StatusFailed
)

const (
	ChainC Chain = iota
	ChainP
)

type TxIssuer interface {
	IssueTx(ctx context.Context, tx *Tx) error
	TrackTx(ctx context.Context, tx *Tx) error
}

type Status int
type Chain int

type Tx struct {
	ReqID string
	TxID  ids.ID
	Chain Chain
	Bytes []byte
	Result
}

type Result struct {
	Status Status
	Reason string
}

type MyTxIssuer struct {
	Logger       logger.Logger
	CChainClient evm.Client
	PChainClient platformvm.Client
}

func (t *MyTxIssuer) IssueTx(ctx context.Context, tx *Tx) error {
	err := backoff.RetryFnExponential10Times(t.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		switch tx.Chain {
		case ChainC:
			_, err := t.CChainClient.IssueTx(ctx, tx.Bytes)
			if err != nil {
				return true, errors.WithStack(err)
			}
		case ChainP:
			_, err := t.PChainClient.IssueTx(ctx, tx.Bytes)
			if err != nil {
				return true, errors.WithStack(err)
			}
		}
		tx.Result = Result{StatusIssued, "Issued"}
		return false, nil
	})

	t.Logger.ErrorOnError(err, "Failed to issue Tx", []logger.Field{{"tx", tx}}...)
	t.Logger.DebugNilError(err, "Issued Tx", []logger.Field{{"tx", tx}}...)
	return errors.WithStack(err)
}

func (t *MyTxIssuer) TrackTx(ctx context.Context, tx *Tx) error {
	err := backoff.RetryFnExponential10Times(t.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		switch tx.Chain {
		case ChainC:
			status, err := t.CChainClient.GetAtomicTxStatus(ctx, tx.TxID)
			if err != nil {
				return true, errors.WithStack(err)
			}

			switch status {
			case evm.Dropped:
				tx.Result = Result{StatusFailed, "Dropped"}
			case evm.Processing:
				tx.Result = Result{StatusProcessing, "Processing"}
			case evm.Accepted:
				tx.Result = Result{StatusApproved, "Accepted"}
			}
		case ChainP:
			statusResp, err := t.PChainClient.GetTxStatus(ctx, tx.TxID)
			if err != nil {
				return true, errors.WithStack(err)
			}
			switch statusResp.Status {
			case status.Aborted:
				tx.Result = Result{StatusFailed, "Aborted"}
			case status.Processing:
				tx.Result = Result{StatusProcessing, "Processing"}
			case status.Dropped:
				tx.Result = Result{StatusFailed, "Dropped"}
			case status.Committed:
				tx.Result = Result{StatusApproved, "Committed"}
			}
		}
		return false, nil
	})

	t.Logger.ErrorOnError(err, "Failed to track Tx", []logger.Field{{"tx", tx}}...)
	t.Logger.DebugNilError(err, "Tracked Tx", []logger.Field{{"tx", tx}}...)
	return errors.WithStack(err)
}
