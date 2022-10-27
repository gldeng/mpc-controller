package core

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	pStatus "github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
)

var (
	_ TaskContext = (*TaskContextImp)(nil)
)

type TaskContextImp struct {
	Logger logger.Logger

	NonceGiver   noncer.Noncer
	Network      chain.NetworkContext
	EthClient    *ethclient.Client
	MpcClient    MpcClient
	CChainClient evm.Client
	PChainClient platformvm.Client
	Dispatcher   kbcevents.Dispatcher[interface{}]
}

type TaskContextImpConfig struct {
	Logger         logger.Logger
	Host           string
	Port           int16
	SslEnabled     bool
	NetworkContext chain.NetworkContext
	MpcClient      MpcClient
}

func (c TaskContextImpConfig) getUri() string {
	scheme := "http"
	if c.SslEnabled {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v:%v", scheme, c.Host, c.Port)
}

func NewTaskContextImp(config TaskContextImpConfig) (*TaskContextImp, error) {
	ethClient, err := ethclient.Dial(fmt.Sprintf("%s/ext/bc/C/rpc", config.getUri()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial eth client")
	}
	cClient := evm.NewClient(config.getUri(), "C")
	pClient := platformvm.NewClient(config.getUri())
	return &TaskContextImp{
		Logger:       config.Logger,
		NonceGiver:   nil,
		Network:      config.NetworkContext,
		EthClient:    ethClient,
		MpcClient:    config.MpcClient,
		CChainClient: cClient,
		PChainClient: pClient,
		Dispatcher:   nil,
	}, err

}

func (t *TaskContextImp) GetLogger() logger.Logger {
	return t.Logger
}

func (t *TaskContextImp) GetNetwork() *chain.NetworkContext {
	return &t.Network
}

func (t *TaskContextImp) GetMpcClient() MpcClient {
	return t.MpcClient
}

func (t *TaskContextImp) IssueCChainTx(txBytes []byte) (ids.ID, error) {
	return t.CChainClient.IssueTx(context.Background(), txBytes)
}

func (t *TaskContextImp) IssuePChainTx(txBytes []byte) (ids.ID, error) {
	return t.PChainClient.IssueTx(context.Background(), txBytes)
}

func (t *TaskContextImp) CheckCChainTx(id ids.ID) (TxStatus, error) {
	status, err := t.CChainClient.GetAtomicTxStatus(context.Background(), id)
	switch status {
	case evm.Unknown:
		return TxStatusUnknown, nil
	case evm.Dropped:
		return TxStatusDropped, nil
	case evm.Processing:
		return TxStatusProcessing, nil
	case evm.Accepted:
		return TxStatusCommitted, nil
	}
	return TxStatusUnknown, err
}

func (t *TaskContextImp) CheckPChainTx(id ids.ID) (TxStatus, error) {
	resp, err := t.PChainClient.GetTxStatus(context.Background(), id)
	switch resp.Status {
	case pStatus.Unknown:
		return TxStatusUnknown, nil
	case pStatus.Committed:
		return TxStatusCommitted, nil
	case pStatus.Aborted:
		return TxStatusAborted, nil
	case pStatus.Processing:
		return TxStatusProcessing, nil
	case pStatus.Dropped:
		return TxStatusDropped, nil
	}
	return TxStatusUnknown, err
}

func (t *TaskContextImp) NonceAt(account common.Address) (uint64, error) {
	return t.EthClient.NonceAt(context.Background(), account, nil)
}

func (t *TaskContextImp) Emit(event interface{}) {
	//TODO implement me
	panic("implement me")
}

func (t *TaskContextImp) GetDb() storage.SlimDb {
	//TODO implement me
	panic("implement me")
}

func (t *TaskContextImp) GetEventID(event string) (common.Hash, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TaskContextImp) GetParticipantID() storage.ParticipantId {
	//TODO implement me
	panic("implement me")
}

func (t *TaskContextImp) Close() {
	t.EthClient.Close()
}

func NewTaskContextImpFactory(config TaskContextImpConfig) (TaskContextFactory, error) {
	_, err := NewTaskContextImp(config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to use the config in TaskContextImpFactory")
	}
	factory := func() TaskContext {
		ctx, _ := NewTaskContextImp(config)
		return ctx
	}
	return factory, err
}
