package chain

import (
	"context"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type PlatformvmClientWrapper struct {
	Logger logger.Logger
	platformvm.Client
}

func (c *PlatformvmClientWrapper) IssueTx(ctx context.Context, tx []byte, options ...rpc.Option) (txID ids.ID, err error) {
	backoff.RetryFnExponentialForever(c.Logger, ctx, time.Millisecond*100, time.Second*10, func() error {
		txID, err = c.Client.IssueTx(ctx, tx, options...)
		if err != nil {
			err = errors.Wrapf(err, "failed to IssueTx with platformvm.Client")
			if strings.Contains(err.Error(), "failed to read consumed UTXO") && strings.Contains(err.Error(), "due to: not found") {
				return nil
			}
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	var resp *platformvm.GetTxStatusResponse
	resp, err = c.AwaitTxDecided(ctx, txID, 1*time.Second)
	if err != nil {
		return
	}
	if resp.Status.String() != "Committed" {
		err = errors.Errorf("transaction failed with platformvm.Client. txID:%q, status:%q, reason:%q", txID, resp.Status.String(), resp.Reason)
	}

	return
}

func (c *PlatformvmClientWrapper) GetTx(ctx context.Context, txID ids.ID, options ...rpc.Option) (resp []byte, err error) {
	backoff.RetryFnExponentialForever(c.Logger, ctx, time.Millisecond*100, time.Second*10, func() error {
		resp, err = c.Client.GetTx(ctx, txID, options...)
		if err != nil {
			err = errors.Wrapf(err, "failed to GetTx with platformvm.Client")
			return err
		}
		return nil
	})
	return
}

func (c *PlatformvmClientWrapper) GetTxStatus(ctx context.Context, txID ids.ID, options ...rpc.Option) (resp *platformvm.GetTxStatusResponse, err error) {
	backoff.RetryFnExponentialForever(c.Logger, ctx, time.Millisecond*100, time.Second*10, func() error {
		resp, err = c.Client.GetTxStatus(ctx, txID, options...)
		if err != nil {
			err = errors.Wrapf(err, "failed to GetTxStatus with platformvm.Client")
			return err
		}
		return nil
	})
	return
}

func (c *PlatformvmClientWrapper) AwaitTxDecided(ctx context.Context, txID ids.ID, freq time.Duration, options ...rpc.Option) (resp *platformvm.GetTxStatusResponse, err error) {
	backoff.RetryFnExponentialForever(c.Logger, ctx, time.Millisecond*100, time.Second*10, func() error {
		resp, err = c.Client.AwaitTxDecided(ctx, txID, freq, options...)
		if err != nil {
			err = errors.Wrapf(err, "failed to AwaitTxDecided with platformvm.Client")
			return err
		}
		return nil
	})
	return
}

func (c *PlatformvmClientWrapper) GetRewardUTXOs(ctx context.Context, args *api.GetTxArgs, opts ...rpc.Option) (utxoBytes [][]byte, err error) {
	backoff.RetryFnExponentialForever(c.Logger, ctx, time.Millisecond*100, time.Second*10, func() error {
		utxoBytes, err = c.Client.GetRewardUTXOs(ctx, args, opts...)
		if err != nil {
			err = errors.Wrapf(err, "failed to get reward UTXOs for txID %q", args.TxID)
			return err
		}

		return nil
	})

	return
}
