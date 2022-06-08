package wrappers

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
)

type MpcManagerTransactorWrapper struct {
	logger.Logger
	*contract.MpcManagerTransactor
}

func (m *MpcManagerTransactorWrapper) ReportGeneratedKey(ctx context.Context, opts *bind.TransactOpts, groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (tx *types.Transaction, err error) {
	err = backoff.RetryFnExponentialForever(m.Logger, ctx, func() error {
		var err error
		tx, err = m.MpcManagerTransactor.ReportGeneratedKey(opts, groupId, myIndex, generatedPublicKey)
		if err != nil {
			m.Error("Failed to report generated key", logger.Field{"error", err})
			return errors.WithStack(err)
		}

		return nil
	})
	return
}

func (m *MpcManagerTransactorWrapper) JoinRequest(ctx context.Context, opts *bind.TransactOpts, requestId *big.Int, myIndex *big.Int) (tx *types.Transaction, err error) {
	err = backoff.RetryFnExponentialForever(m.Logger, ctx, func() error {
		var err error
		tx, err = m.MpcManagerTransactor.JoinRequest(opts, requestId, myIndex)
		if err != nil {
			m.Error("Failed to join request", logger.Field{"error", err})
			return errors.WithStack(err)
		}

		return nil
	})
	return
}
