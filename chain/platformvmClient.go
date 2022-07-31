package chain

import (
	"context"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	mpcErrors "github.com/avalido/mpc-controller/utils/errors"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type PlatformvmClientWrapper struct {
	Logger logger.Logger
	platformvm.Client
}

func (c *PlatformvmClientWrapper) IssueTx(ctx context.Context, tx []byte, options ...rpc.Option) (txID ids.ID, err error) {
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		txID, err = c.Client.IssueTx(ctx, tx, options...)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "shared memory: not found"):
				return false, errors.WithStack(mpcErrors.Wrap(err, &ErrTypSharedMemoryNotFound{}))
			case strings.Contains(errMsg, "consumed UTXO not found") || strings.Contains(errMsg, "failed to read consumed UTXO"):
				return false, errors.WithStack(mpcErrors.Wrap(err, &ErrTypConsumedUTXONotFound{}))
			case strings.Contains(errMsg, "not before validator's start time") || strings.Contains(errMsg, "later than staker start time"):
				return false, errors.WithStack(mpcErrors.Wrap(err, &ErrTypStakeStartTimeExpired{}))
			default:
				return true, errors.WithStack(err) // todo: exploring more concrete error types, including connection failure
			}
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to Issue tx")
	if err != nil {
		return
	}

	var resp *platformvm.GetTxStatusResponse
	resp, err = c.AwaitTxDecided(ctx, txID, 1*time.Second)
	if err != nil {
		return
	}
	if resp.Status.String() != "Committed" {
		err = errors.Errorf("issued transaction failed. txID:%q, status:%q, reason:%q", txID, resp.Status.String(), resp.Reason)
	}
	return
}

func (c *PlatformvmClientWrapper) GetTx(ctx context.Context, txID ids.ID, options ...rpc.Option) (resp []byte, err error) {
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		resp, err = c.Client.GetTx(ctx, txID, options...)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get tx")
	return
}

func (c *PlatformvmClientWrapper) GetTxStatus(ctx context.Context, txID ids.ID, options ...rpc.Option) (resp *platformvm.GetTxStatusResponse, err error) {
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		resp, err = c.Client.GetTxStatus(ctx, txID, options...)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get tx status")
	return
}

func (c *PlatformvmClientWrapper) AwaitTxDecided(ctx context.Context, txID ids.ID, freq time.Duration, options ...rpc.Option) (resp *platformvm.GetTxStatusResponse, err error) {
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		resp, err = c.Client.AwaitTxDecided(ctx, txID, freq, options...)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to await tx decided")
	return
}

func (c *PlatformvmClientWrapper) GetRewardUTXOs(ctx context.Context, args *api.GetTxArgs, opts ...rpc.Option) (utxoBytes [][]byte, err error) {
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		utxoBytes, err = c.Client.GetRewardUTXOs(ctx, args, opts...)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get reward UTXOs for txID %q", args.TxID)
	return
}
