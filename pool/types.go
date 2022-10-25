package pool

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	pStatus "github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/chain"
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
	IsDone() bool
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

type TaskContextImp struct {
	Logger logger.Logger

	NonceGiver   noncer.Noncer
	Network      chain.NetworkContext
	EthClient    *ethclient.Client
	MpcClient    core.MpcClient
	CChainClient evm.Client
	PChainClient platformvm.Client
	Dispatcher   kbcevents.Dispatcher[interface{}]
}

func (t *TaskContextImp) GetLogger() logger.Logger {
	return t.Logger
}

func (t *TaskContextImp) GetNetwork() *chain.NetworkContext {
	return &t.Network
}

func (t *TaskContextImp) GetMpcClient() core.MpcClient {
	return t.MpcClient
}

func (t *TaskContextImp) IssueCChainTx(txBytes []byte) (ids.ID, error) {
	return t.CChainClient.IssueTx(context.Background(), txBytes)
}

func (t *TaskContextImp) IssuePChainTx(txBytes []byte) (ids.ID, error) {
	return t.PChainClient.IssueTx(context.Background(), txBytes)
}

func (t *TaskContextImp) CheckCChainTx(id ids.ID) (Status, error) {
	status, err := t.CChainClient.GetAtomicTxStatus(context.Background(), id)
	switch status {
	case evm.Unknown:
		return Unknown, nil
	case evm.Dropped:
		return Dropped, nil
	case evm.Processing:
		return Processing, nil
	}
	return Unknown, err
}

func (t *TaskContextImp) CheckPChainTx(id ids.ID) (Status, error) {
	resp, err := t.PChainClient.GetTxStatus(context.Background(), id)
	switch resp.Status {
	case pStatus.Unknown:
		return Unknown, nil
	case pStatus.Committed:
		return Committed, nil
	case pStatus.Aborted:
		return Aborted, nil
	case pStatus.Processing:
		return Processing, nil
	case pStatus.Dropped:
		return Dropped, nil
	}
	return Unknown, err
}

func (t *TaskContextImp) NonceAt(account common.Address) (uint64, error) {
	return t.EthClient.NonceAt(context.Background(), account, nil)
}

func (t *TaskContextImp) Emit(event interface{}) {
	//TODO implement me
	panic("implement me")
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
