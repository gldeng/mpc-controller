package chain

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding chain status

type Receipter interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (r *types.Receipt, err error)
}

type Noncer interface {
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (nonce uint64, err error)
}

type Balancer interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (bl *big.Int, err error)
}

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding issue Tx

type CChainIssuer interface {
	IssueTx(ctx context.Context, txBytes []byte) (ids.ID, error)
}

type PChainIssuer interface {
	IssueTx(ctx context.Context, tx []byte, options ...rpc.Option) (ids.ID, error)
}
