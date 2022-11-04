package core

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	pStatus "github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

	Services     *ServicePack
	NonceGiver   noncer.Noncer
	Network      chain.NetworkContext
	EthClient    *ethclient.Client
	MpcClient    MpcClient
	CChainClient evm.Client
	PChainClient platformvm.Client
	Db           storage.SlimDb
	abi          *abi.ABI
	Dispatcher   kbcevents.Dispatcher[interface{}]
}

func (t *TaskContextImp) CheckEthTx(txHash common.Hash) (TxStatus, error) {
	rcp, err := t.EthClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return TxStatusUnknown, errors.Wrap(err, "failed to check tx receipt")
	}
	if rcp.Status == 0 {
		return TxStatusAborted, nil
	}
	if rcp.Status == 1 {
		return TxStatusCommitted, nil
	}
	return TxStatusUnknown, errors.New(fmt.Sprintf("unknown tx status %v", rcp.Status))
}

func (t *TaskContextImp) ReportGeneratedKey(opts *bind.TransactOpts, participantId [32]byte, generatedPublicKey []byte) (*common.Hash, error) {
	transactor, err := contract.NewMpcManagerTransactor(t.Services.Config.MpcManagerAddress, t.EthClient)

	if err != nil {
		return nil, errors.Wrap(err, "failed to MpcManagerTransactor")
	}
	tx, err := transactor.ReportGeneratedKey(opts, participantId, generatedPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed send tx")
	}
	hash := tx.Hash()
	return &hash, nil
}

func (t *TaskContextImp) GetGroup(opts *bind.CallOpts, groupId [32]byte) ([][]byte, error) {
	caller, err := contract.NewMpcManagerCaller(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MpcManagerCaller")
	}
	return caller.GetGroup(opts, groupId)
}

func NewTaskContextImp(services *ServicePack) (*TaskContextImp, error) {
	ethClient := services.Config.CreateEthClient()
	cClient := services.Config.CreateCClient()
	pClient := services.Config.CreatePClient()
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get abi")
	}
	return &TaskContextImp{
		Logger:       services.Logger,
		Services:     services,
		NonceGiver:   nil,
		Network:      services.Config.NetworkContext,
		EthClient:    ethClient,
		MpcClient:    services.MpcClient,
		CChainClient: cClient,
		PChainClient: pClient,
		Db:           services.Db,
		abi:          abi,
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
	if err != nil {
		return TxStatusUnknown, errors.Wrapf(err, "failed to get C-Chain AtomicTx status for %v", id)
	}
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
	if err != nil {
		return TxStatusUnknown, errors.Wrapf(err, "failed to get P-Chain Tx status for %v", id)
	}
	if resp == nil {
		return TxStatusUnknown, errors.Errorf("got nil P-Chain TxStatusResponse for %v", id)
	}
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
	return t.Db
}

func (t *TaskContextImp) GetEventID(event string) (common.Hash, error) {
	return t.abi.Events[event].ID, nil
}

func (t *TaskContextImp) LoadGroup(groupID [32]byte) (*types.Group, error) {
	key := []byte("group/")
	key = append(key, groupID[:]...)
	groupBytes, err := t.Db.Get(context.Background(), key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load group")
	}

	group := &types.Group{}
	err = group.Decode(groupBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode group")
	}
	return group, nil
}

func (t *TaskContextImp) GetParticipantID() storage.ParticipantId {
	//TODO Do it properly
	id, err := t.Db.Get(context.Background(), []byte("participant_id"))
	if err != nil {
		panic(err)
	}
	var id32 [32]byte
	copy(id32[:], id)
	return id32
}

func (t *TaskContextImp) GetMyPublicKey() ([]byte, error) {
	return t.Services.Config.MyPublicKey, nil
}

func (t *TaskContextImp) GetMyTransactSigner() *bind.TransactOpts {
	return t.Services.Config.MyTransactSigner
}

func (t *TaskContextImp) Close() {
	t.EthClient.Close()
}

func NewTaskContextImpFactory(services *ServicePack) (TaskContextFactory, error) {
	_, err := NewTaskContextImp(services)
	if err != nil {
		return nil, errors.Wrap(err, "unable to use the config in TaskContextImpFactory")
	}
	factory := func() TaskContext {
		ctx, _ := NewTaskContextImp(services)
		return ctx
	}
	return factory, err
}
