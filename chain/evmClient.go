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
	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		txID, err = c.Client.IssueTx(ctx, txBytes)
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, ErrMsgInsufficientFunds):
				return false, errors.WithStack(&ErrTypInsufficientFunds{Cause: err})
			case strings.Contains(errMsg, ErrMsgInvalidNonce):
				return false, errors.WithStack(&ErrTypInvalidNonce{Cause: err})
			case strings.Contains(errMsg, ErrMsgConflictAtomicInputs):
				return false, errors.WithStack(&ErrTypConflictAtomicInputs{Cause: err})
			case strings.Contains(errMsg, ErrMsgTxHasNoImportedInputs):
				return false, errors.WithStack(&ErrTypTxHasNoImportedInputs{Cause: err})
			case strings.Contains(errMsg, "failed to fetch import UTXOs"):
				return false, errors.WithStack(&ErrTypImportUTXOsNotFound{Cause: err})
			case strings.Contains(errMsg, ErrMsgNotFound):
				return false, errors.WithStack(&ErrTypNotFound{Cause: err})
			default:
				return true, errors.WithStack(err) // todo: exploring more concrete error types, including connection failure
			}
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
	err = backoff.RetryFnExponential10Times(c.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		status, err = c.Client.GetAtomicTxStatus(ctx, txID)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get atomic tx status")
	return
}
