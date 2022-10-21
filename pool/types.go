package pool

import (
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	kbcevents "github.com/kubecost/events"
)

// TODO: Deprecate
type WorkerPool interface {
	Submit(task func())
	StopAndWait()
}

type Task interface {
	Next(ctx TaskContext) ([]Task, error)
}

type TaskContextImp struct { // TODO: Convert it to TaskApi interface instead of directly giving the underlying resources
	Logger logger.Logger

	NonceGiver noncer.Noncer
	Network    chain.NetworkContext
	EthClient  *ethclient.Client
	MpcClient  core.MpcClient
	TxIssuer   txissuer.TxIssuer
	Dispatcher kbcevents.Dispatcher[interface{}]
}

type TaskContext interface {
	GetLogger() logger.Logger
	GetNetwork() *chain.NetworkContext
	GetMpcClient() core.MpcClient
	GetTxIssuer() txissuer.TxIssuer
	NonceAt(account common.Address) (uint64, error)
	Emit(event interface{})
}

type TaskContextFactory = func() TaskContext
type TaskSubmitter interface {
	Submit(task Task) error
}
