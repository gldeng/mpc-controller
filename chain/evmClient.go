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

func (c *EvmClientWrapper) IssueTx(ctx context.Context, txBytes []byte) (txID ids.ID, err error) {
	backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() error {
		txID, err = c.Client.IssueTx(ctx, txBytes)
		if err != nil {
			err = errors.Wrapf(err, "failed to IssueTx with evm.Client")
			if strings.Contains(err.Error(), "insufficient funds") ||
				strings.Contains(err.Error(), "tx has no imported inputs") ||
				strings.Contains(err.Error(), "invalid nonce") ||
				strings.Contains(err.Error(), "invalid block due to conflicting atomic inputs") ||
				strings.Contains(err.Error(), "due to: not found") {
				return nil
			}
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	var status evm.Status
	status, err = c.AwaitTxDecided(ctx, txID, 1*time.Second)
	if err != nil {
		return
	}
	if status != evm.Accepted {
		err = errors.Errorf("transaction failed with evm.Client. txID:%q, status:%q", txID, status)
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
	backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() error {
		status, err = c.Client.GetAtomicTxStatus(ctx, txID)
		if err != nil {
			err = errors.Wrapf(err, "failed to GetAtomicTxStatus with evm.Client")
			return err
		}
		return nil
	})
	return
}
