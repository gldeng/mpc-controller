package core

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
)

type Task interface {
	Next(ctx TaskContext) ([]Task, error)
	IsDone() bool
	RequiresNonce() bool
}

type TaskContext interface {
	GetLogger() logger.Logger
	GetNetwork() *chain.NetworkContext
	GetMpcClient() MpcClient
	IssueCChainTx(txBytes []byte) (ids.ID, error)
	IssuePChainTx(txBytes []byte) (ids.ID, error)
	CheckCChainTx(id ids.ID) (TxStatus, error)
	CheckPChainTx(id ids.ID) (TxStatus, error)
	NonceAt(account common.Address) (uint64, error)
	Emit(event interface{})
}

type TaskContextFactory = func() TaskContext
type TaskSubmitter interface {
	Submit(task Task) error
}
