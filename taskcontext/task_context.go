package taskcontext

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	pStatus "github.com/ava-labs/avalanchego/vms/platformvm/status"
	coreTypes "github.com/ava-labs/coreth/core/types"
	"github.com/ava-labs/coreth/core/vm"
	"github.com/ava-labs/coreth/interfaces"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/mpc"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"math/big"
	"strings"
)

var (
	_ core.TaskContext = (*TaskContextImp)(nil)
)

type TaskContextImp struct {
	Logger logger.Logger

	Services     *core.ServicePack
	NonceGiver   noncer.Noncer
	Network      core.NetworkContext
	EthClient    *ethclient.Client // TODO: use ava-labs/coreth/ethclient instead for future 100% compatibility
	MpcClient    mpc.MpcClient
	CChainClient evm.Client
	PChainClient platformvm.Client
	Db           core.Store
	abi          *abi.ABI
	Dispatcher   kbcevents.Dispatcher[interface{}]
}

// Reference: https://github.com/ethereum/go-ethereum/blob/v1.10.26/core/types/receipt.go
// Reference: https://github.com/ava-labs/coreth/blob/v0.8.13/interfaces/interfaces.go

func (t *TaskContextImp) CheckEthTx(txHash common.Hash) (core.TxStatus, error) {
	rcp, err := t.EthClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		if strings.Contains(err.Error(), interfaces.NotFound.Error()) {
			return core.TxStatusUnknown, errors.WithStack(&ErrTypTxStatusNotFound{Pre: err})
		}
		return core.TxStatusUnknown, errors.WithStack(&ErrTypTxStatusQueryFail{Pre: err})
	}

	if rcp.Status == coreTypes.ReceiptStatusFailed {
		return core.TxStatusAborted, errors.WithStack(ErrTxStatusAborted)
	}
	return core.TxStatusCommitted, nil
}

func (t *TaskContextImp) ReportGeneratedKey(opts *bind.TransactOpts, participantId [32]byte, generatedPublicKey []byte) (*common.Hash, error) {
	transactor, err := contract.NewMpcManagerTransactor(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.WithStack(&ErrTypContractBindFail{Pre: err})
	}

	var hash common.Hash
	tx, err := transactor.ReportGeneratedKey(opts, participantId, generatedPublicKey)
	if err != nil { // reference: https://github.com/ava-labs/coreth/blob/v0.8.13/core/vm/errors.go
		if strings.Contains(err.Error(), vm.ErrExecutionReverted.Error()) { // TODO: find and find more errors: https://github.com/AvaLido/mpc-controller/issues/156
			return nil, errors.WithStack(&ErrTypTxReverted{Pre: err})
		}
		return nil, errors.WithStack(&ErrTypContractCallFail{Pre: err})
	}

	hash = tx.Hash()
	return &hash, nil
}

func (t *TaskContextImp) JoinRequest(opts *bind.TransactOpts, participantId [32]byte, requestHash [32]byte) (*common.Hash, error) {
	transactor, err := contract.NewMpcManagerTransactor(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.WithStack(&ErrTypContractBindFail{Pre: err})
	}

	var hash common.Hash
	tx, err := transactor.JoinRequest(opts, participantId, requestHash)
	if err != nil { // reference: https://github.com/ava-labs/coreth/blob/v0.8.13/core/vm/errors.go
		if strings.Contains(err.Error(), vm.ErrExecutionReverted.Error()) { // TODO: find and find more errors: https://github.com/AvaLido/mpc-controller/issues/156
			return nil, errors.WithStack(&ErrTypTxReverted{Pre: err})
		}
		return nil, errors.WithStack(&ErrTypContractCallFail{Pre: err})
	}

	hash = tx.Hash()
	return &hash, nil
}

func (t *TaskContextImp) GetGroup(opts *bind.CallOpts, groupId [32]byte) ([][]byte, error) {
	caller, err := contract.NewMpcManagerCaller(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MpcManagerCaller")
	}
	return caller.GetGroup(opts, groupId)
}

func (t *TaskContextImp) LastGenPubKey(opts *bind.CallOpts) ([]byte, error) {
	caller, err := contract.NewMpcManagerCaller(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MpcManagerCaller")
	}
	return caller.LastGenPubKey(opts)
}

func (t *TaskContextImp) GetGroupIdByKey(opts *bind.CallOpts, publicKey []byte) ([32]byte, error) {
	caller, err := contract.NewMpcManagerCaller(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "failed to create MpcManagerCaller")
	}
	return caller.GetGroupIdByKey(opts, publicKey)
}

func (t *TaskContextImp) RequestConfirmations(opts *bind.CallOpts, groupId [32]byte, requestHash [32]byte) (*big.Int, error) {
	caller, err := contract.NewMpcManagerCaller(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MpcManagerCaller")
	}
	return caller.RequestRecords(opts, groupId, requestHash)
}

func NewTaskContextImp(services *core.ServicePack) (*TaskContextImp, error) {
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

func (t *TaskContextImp) GetNetwork() *core.NetworkContext {
	return &t.Network
}

func (t *TaskContextImp) GetMpcClient() mpc.MpcClient {
	return t.MpcClient
}

func (t *TaskContextImp) IssueCChainTx(txBytes []byte) (ids.ID, error) {
	id, err := t.CChainClient.IssueTx(context.Background(), txBytes)
	if err != nil { // reference: https://github.com/ava-labs/coreth/blob/v0.8.13/plugin/evm/vm.go#L147
		switch { // TODO: find and find more errors: https://github.com/AvaLido/mpc-controller/issues/156
		case strings.Contains(err.Error(), "insufficient funds"):
			fallthrough
		case strings.Contains(err.Error(), "invalid nonce"):
			fallthrough
		case strings.Contains(err.Error(), "conflicting atomic inputs"):
			fallthrough
		case strings.Contains(err.Error(), "tx has no imported inputs"):
			return id, errors.WithStack(&ErrTypeTxInputsInvalid{Pre: err})
		case strings.Contains(err.Error(), "import UTXOs not found"):
			return id, errors.WithStack(&ErrTypeUTXOConsumeFail{Pre: err})
		}
		return id, errors.WithStack(&ErrTypeTxIssueFail{Pre: err})
	}
	return id, nil
}

func (t *TaskContextImp) IssuePChainTx(txBytes []byte) (ids.ID, error) {
	id, err := t.PChainClient.IssueTx(context.Background(), txBytes)
	if err != nil {
		switch { // TODO: find and handle more errors: https://github.com/AvaLido/mpc-controller/issues/156
		case strings.Contains(err.Error(), "failed to get shared memory"): // reference: https://github.com/ava-labs/avalanchego/blob/v1.7.14/vms/platformvm/standard_tx_executor.go#L158
			return id, errors.WithStack(&ErrTypSharedMemoryFail{Pre: err})
		case strings.Contains(err.Error(), "failed to read consumed UTXO"): // reference: https://github.com/ava-labs/avalanchego/blob/v1.7.14/vms/platformvm/utxo/handler.go#L458
			return id, errors.WithStack(&ErrTypeUTXOConsumeFail{Pre: err})
		case strings.Contains(err.Error(), "not before validator's start time"): // reference: https://github.com/ava-labs/avalanchego/blob/v1.7.14/vms/platformvm/proposal_tx_executor.go#L380
			return id, errors.WithStack(&ErrTypeTxInputsInvalid{Pre: err})
		}
		return id, errors.WithStack(&ErrTypeTxIssueFail{Pre: err})
	}
	return id, nil
}

// Reference: https://github.com/ava-labs/coreth/blob/v0.8.13/plugin/evm/status.go

func (t *TaskContextImp) CheckCChainTx(id ids.ID) (core.TxStatus, error) {
	status, err := t.CChainClient.GetAtomicTxStatus(context.Background(), id)
	if err != nil {
		return core.TxStatusUnknown, errors.WithStack(&ErrTypTxStatusQueryFail{Pre: err})
	}

	switch status {
	case evm.Unknown:
		return core.TxStatusUnknown, nil
	case evm.Accepted:
		return core.TxStatusCommitted, nil
	case evm.Processing:
		return core.TxStatusProcessing, nil
	case evm.Dropped:
		return core.TxStatusDropped, nil
	default:
		return core.TxStatusInvalid, nil
	}
}

// Reference: https://github.com/ava-labs/avalanchego/blob/v1.7.14/vms/platformvm/status/status.go

func (t *TaskContextImp) CheckPChainTx(id ids.ID) (core.Status, error) {
	resp, err := t.PChainClient.GetTxStatus(context.Background(), id)
	if err != nil {
		return core.Status{core.TxStatusUnknown, "failed to query status"}, errors.WithStack(&ErrTypTxStatusQueryFail{Pre: err})
	}

	switch resp.Status {
	case pStatus.Unknown:
		return core.Status{core.TxStatusUnknown, resp.Reason}, nil
	case pStatus.Committed:
		return core.Status{core.TxStatusCommitted, resp.Reason}, nil
	case pStatus.Aborted:
		return core.Status{core.TxStatusAborted, resp.Reason}, nil
	case pStatus.Processing:
		return core.Status{core.TxStatusProcessing, resp.Reason}, nil
	case pStatus.Dropped:
		return core.Status{core.TxStatusDropped, resp.Reason}, nil
	default:
		return core.Status{core.TxStatusInvalid, resp.Reason}, nil
	}
}

func (t *TaskContextImp) NonceAt(account common.Address) (uint64, error) {
	return t.EthClient.NonceAt(context.Background(), account, nil)
}

func (t *TaskContextImp) Emit(event interface{}) {
	//TODO implement me
	panic("implement me")
}

func (t *TaskContextImp) GetDb() core.Store {
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
		return nil, errors.Wrapf(err, "failed to decode group: %v %v", key, groupBytes)
	}
	return group, nil
}

func (t *TaskContextImp) LoadGroupByLatestMpcPubKey() (*types.Group, error) {
	bytes, err := t.Db.Get(context.Background(), []byte("latestPubKey"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load latest mpc public key")
	}

	if bytes == nil {
		return nil, errors.New("loaded empty mpc public key")
	}

	model := types.MpcPublicKey{}
	err = model.Decode(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode mpc public key")
	}

	group, err := t.LoadGroup(model.GroupId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load group")
	}

	return group, nil
}

func (t *TaskContextImp) GetParticipantID() types.ParticipantId {
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

func NewTaskContextImpFactory(services *core.ServicePack) (core.TaskContextFactory, error) {
	_, err := NewTaskContextImp(services)
	if err != nil {
		return nil, errors.Wrap(err, "unable to use the config in TaskContextImpFactory")
	}
	factory := func() core.TaskContext {
		ctx, _ := NewTaskContextImp(services)
		return ctx
	}
	return factory, err
}
