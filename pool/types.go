package pool

import (
	"github.com/ava-labs/avalanchego/ids"
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

type Status = int

const (
	Unknown    Status = 0
	Committed  Status = 4
	Aborted    Status = 5
	Processing Status = 6
	Dropped    Status = 8
)

func IsPending(status Status) bool {
	if status == Unknown {
		return true
	}
	if status == Processing {
		return true
	}
	return false
}

func IsSuccessful(status Status) bool {
	if status == Committed {
		return true
	}
	return false
}

func IsFailed(status Status) bool {
	if status == Aborted {
		return true
	}
	if status == Dropped {
		return true
	}
	return false
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
	IssueCChainTx(txBytes []byte) (ids.ID, error)
	IssuePChainTx(txBytes []byte) (ids.ID, error)
	CheckCChainTx(id ids.ID) (Status, error)
	CheckPChainTx(id ids.ID) (Status, error)
	NonceAt(account common.Address) (uint64, error)
	Emit(event interface{})
}

type TaskContextFactory = func() TaskContext
type TaskSubmitter interface {
	Submit(task Task) error
}
