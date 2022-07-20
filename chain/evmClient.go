package chain

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type EvmClientWrapper struct {
	Logger logger.Logger
	evm.Client
}

// todo: define error types
// todo: further strategy to handle errors, e.g., check SPI request
// SPC: short of 'single participant computation', referring to those activity that allow only one success execution
// e.g. issue atomic tx after MPC signing stage

func (c *EvmClientWrapper) IssueTx(ctx context.Context, txBytes []byte) (txID ids.ID, err error) {
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		txID, err = c.Client.IssueTx(ctx, txBytes)
		if err != nil {
			if strings.Contains(err.Error(), "insufficient funds") ||
				strings.Contains(err.Error(), "tx has no imported inputs") ||
				strings.Contains(err.Error(), "invalid nonce") ||
				strings.Contains(err.Error(), "invalid block due to conflicting atomic inputs") ||
				strings.Contains(err.Error(), "due to: not found") {
				return false, err
			}
			return true, err
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to issue tx")
	if err != nil {
		return
	}

	var status evm.Status
	status, err = c.AwaitTxDecided(ctx, txID, 1*time.Second)
	if err != nil {
		err = errors.Wrapf(err, "failed to await tx decided, txID:%q", txID)
		return
	}
	if status != evm.Accepted {
		err = errors.Errorf("issued tx failed. txID:%q, status:%q", txID, status)
	}
	return
}

func (c *EvmClientWrapper) AwaitTxDecided(ctx context.Context, txID ids.ID, freq time.Duration) (status evm.Status, err error) {
	for {
		status, err = c.GetAtomicTxStatus(ctx, txID)
		if err != nil {
			break
		}
		if status == evm.Processing {
			time.Sleep(freq)
			continue
		}
		break
	}
	return
}

func (c *EvmClientWrapper) GetAtomicTxStatus(ctx context.Context, txID ids.ID) (status evm.Status, err error) {
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		status, err = c.Client.GetAtomicTxStatus(ctx, txID)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get atomic tx status")
	return
}
