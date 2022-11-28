package taskcontext

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"math/big"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	pStatus "github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
)

var (
	_ core.TaskContext = (*TaskContextImp)(nil)
)

type TaskContextImp struct {
	Logger logger.Logger

	Services     *core.ServicePack
	NonceGiver   noncer.Noncer
	Network      core.NetworkContext
	EthClient    *ethclient.Client
	MpcClient    core.MpcClient
	CChainClient evm.Client
	PChainClient platformvm.Client
	Db           core.Store
	abi          *abi.ABI
	Dispatcher   kbcevents.Dispatcher[interface{}]
}

func (t *TaskContextImp) CheckEthTx(txHash common.Hash) (core.TxStatus, error) {
	//rcp, err := t.EthClient.TransactionReceipt(context.Background(), txHash)
	//if err != nil {
	//	return TxStatusUnknown, errors.Wrap(err, "failed to check tx receipt")
	//}
	//if rcp.Status == 0 {
	//	return TxStatusAborted, nil
	//}
	//if rcp.Status == 1 {
	//	return TxStatusCommitted, nil
	//}
	//return TxStatusUnknown, errors.New(fmt.Sprintf("unknown tx status %v", rcp.Status))

	var txStatus core.TxStatus
	var rcp *types2.Receipt
	var err error
	err = backoff.RetryFnExponential10Times(t.Logger, context.Background(), time.Second, time.Second*10, func() (retry bool, err error) {
		rcp, err = t.EthClient.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return true, nil // NOTE: tx already sent may take some time to process before being able to check by client
			}
			return true, errors.Wrapf(err, "failed to check receipt for tx %x", txHash)
		}

		if rcp.Status != 1 {
			return true, nil // NOTE: tx maybe in intermediate status thus it makes sense to recheck after some time
		}

		return false, nil
	})

	if err != nil {
		return core.TxStatusUnknown, errors.WithStack(err)
	}

	txStatus = core.TxStatusUnknown
	switch rcp.Status {
	case 0:
		txStatus = core.TxStatusAborted
	case 1:
		txStatus = core.TxStatusCommitted
	}

	return txStatus, nil
}

func (t *TaskContextImp) ReportGeneratedKey(opts *bind.TransactOpts, participantId [32]byte, generatedPublicKey []byte) (*common.Hash, error) {
	//transactor, err := contract.NewMpcManagerTransactor(t.Services.Config.MpcManagerAddress, t.EthClient)
	//
	//if err != nil {
	//	return nil, errors.Wrap(err, "failed to MpcManagerTransactor")
	//}
	//tx, err := transactor.ReportGeneratedKey(opts, participantId, generatedPublicKey)
	//if err != nil {
	//	return nil, errors.Wrap(err, "failed send tx")
	//}
	//hash := tx.Hash()
	//return &hash, nil

	transactor, err := contract.NewMpcManagerTransactor(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MpcManagerTransactor")
	}
	var hash common.Hash
	err = backoff.RetryFnExponential10Times(t.Logger, context.Background(), time.Second, time.Second*10, func() (bool, error) {
		tx, err := transactor.ReportGeneratedKey(opts, participantId, generatedPublicKey)
		if err != nil {
			if strings.Contains(err.Error(), "execution reverted") {
				return false, errors.Wrap(err, "ReportGeneratedKey reverted")
			}
			return true, errors.Wrap(err, "failed call ReportGeneratedKey")
		}
		hash = tx.Hash()
		return false, nil
	})

	return &hash, errors.WithStack(err)
}

func (t *TaskContextImp) JoinRequest(opts *bind.TransactOpts, participantId [32]byte, requestHash [32]byte) (*common.Hash, error) {
	//transactor, err := contract.NewMpcManagerTransactor(t.Services.Config.MpcManagerAddress, t.EthClient)
	//
	//if err != nil {
	//	return nil, errors.Wrap(err, "failed to MpcManagerTransactor")
	//}
	//tx, err := transactor.JoinRequest(opts, participantId, requestHash)
	//if err != nil {
	//	return nil, errors.Wrap(err, "failed send tx")
	//}
	//hash := tx.Hash()
	//return &hash, nil

	transactor, err := contract.NewMpcManagerTransactor(t.Services.Config.MpcManagerAddress, t.EthClient)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create MpcManagerTransactor")
	}
	var hash common.Hash
	err = backoff.RetryFnExponential10Times(t.Logger, context.Background(), time.Second, time.Second*10, func() (bool, error) {
		tx, err := transactor.JoinRequest(opts, participantId, requestHash)
		if err != nil {
			if strings.Contains(err.Error(), "execution reverted") {
				return false, errors.Wrap(err, "JoinRequest reverted")
			}
			return true, errors.Wrap(err, "failed call JoinRequest")
		}
		hash = tx.Hash()
		return false, nil
	})

	return &hash, errors.WithStack(err)
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
	return caller.RequestConfirmations(opts, groupId, requestHash)
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

func (t *TaskContextImp) GetMpcClient() core.MpcClient {
	return t.MpcClient
}

func (t *TaskContextImp) IssueCChainTx(txBytes []byte) (ids.ID, error) {
	return t.CChainClient.IssueTx(context.Background(), txBytes)
}

func (t *TaskContextImp) IssuePChainTx(txBytes []byte) (ids.ID, error) {
	return t.PChainClient.IssueTx(context.Background(), txBytes)
}

func (t *TaskContextImp) CheckCChainTx(id ids.ID) (core.TxStatus, error) {
	status, err := t.CChainClient.GetAtomicTxStatus(context.Background(), id)
	if err != nil {
		return core.TxStatusUnknown, errors.Wrapf(err, "failed to get C-Chain AtomicTx status for %v", id)
	}
	switch status {
	case evm.Unknown:
		return core.TxStatusUnknown, nil
	case evm.Dropped:
		return core.TxStatusDropped, nil
	case evm.Processing:
		return core.TxStatusProcessing, nil
	case evm.Accepted:
		return core.TxStatusCommitted, nil
	}
	return core.TxStatusUnknown, err
}

func (t *TaskContextImp) CheckPChainTx(id ids.ID) (core.TxStatus, error) {
	resp, err := t.PChainClient.GetTxStatus(context.Background(), id)
	if err != nil {
		return core.TxStatusUnknown, errors.Wrapf(err, "failed to get P-Chain Tx status for %v", id)
	}
	if resp == nil {
		return core.TxStatusUnknown, errors.Errorf("got nil P-Chain TxStatusResponse for %v", id)
	}
	switch resp.Status {
	case pStatus.Unknown:
		return core.TxStatusUnknown, nil
	case pStatus.Committed:
		return core.TxStatusCommitted, nil
	case pStatus.Aborted:
		return core.TxStatusAborted, nil
	case pStatus.Processing:
		return core.TxStatusProcessing, nil
	case pStatus.Dropped:
		return core.TxStatusDropped, nil
	}
	return core.TxStatusUnknown, err
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
