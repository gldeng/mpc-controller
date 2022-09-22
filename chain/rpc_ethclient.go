package chain

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

type RpcEthClientWrapper struct {
	logger.Logger
	*ethclient.Client
}

func (m *RpcEthClientWrapper) TransactionReceipt(ctx context.Context, txHash common.Hash) (r *types.Receipt, err error) {
	err = backoff.RetryFnExponentialForever(m.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		r, err = m.Client.TransactionReceipt(ctx, txHash)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to query transaction receipt")
	return
}

func (m *RpcEthClientWrapper) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (nonce uint64, err error) {
	err = backoff.RetryFnExponentialForever(m.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		nonce, err = m.Client.NonceAt(ctx, account, blockNumber)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to query nonce")
	return
}

func (m *RpcEthClientWrapper) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (bl *big.Int, err error) {
	err = backoff.RetryFnExponentialForever(m.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
		bl, err = m.Client.BalanceAt(ctx, account, blockNumber)
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to query balance")
	return
}
