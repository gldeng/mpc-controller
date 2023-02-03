package core

import (
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/core/mpc"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Task interface {
	GetId() string
	Next(ctx TaskContext) ([]Task, error)
	IsDone() bool
	FailedPermanently() bool
	IsSequential() bool
}

type MpcManager interface {
	GetGroup(opts *bind.CallOpts, groupId [32]byte) ([][]byte, error)
	ReportGeneratedKey(opts *bind.TransactOpts, participantId [32]byte, generatedPublicKey []byte) (*common.Hash, error)
	JoinRequest(opts *bind.TransactOpts, participantId [32]byte, requestHash [32]byte) (*common.Hash, error)
	GetGroupIdByKey(opts *bind.CallOpts, publicKey []byte) ([32]byte, error)
	RequestRecords(opts *bind.CallOpts, groupId [32]byte, requestHash [32]byte) (*big.Int, error)
	LastGenPubKey(opts *bind.CallOpts) ([]byte, error)
}

type TaskContext interface {
	MpcManager
	GetLogger() logger.Logger
	GetNetwork() *NetworkContext
	GetMpcClient() mpc.MpcClient
	GetTxIndex() TxIndexReader
	IssueCChainTx(txBytes []byte) (ids.ID, error)
	IssuePChainTx(txBytes []byte) (ids.ID, error)
	CheckEthTx(txHash common.Hash) (TxStatus, error)
	CheckCChainTx(id ids.ID) (TxStatus, error)
	CheckPChainTx(id ids.ID) (Status, error)
	GetPChainTx(txID ids.ID) (*txs.Tx, error)
	NonceAt(account common.Address) (uint64, error)
	Emit(event interface{})
	GetDb() Store
	GetEventID(event string) (common.Hash, error)
	GetMyPublicKey() ([]byte, error)
	GetMyTransactSigner() *bind.TransactOpts
	LoadGroup(groupID [32]byte) (*types.Group, error)
	LoadGroupByLatestMpcPubKey() (*types.Group, error)
	GetParticipantID() types.ParticipantId
}

type TaskContextFactory = func() TaskContext
type TaskSubmitter interface {
	Submit(task Task) error
}

// FlowId is used to link multiple tasks belonging to the same unit of work
type FlowId struct {
	Tag         string
	RequestHash types.RequestHash
}

func (id FlowId) String() string {
	return fmt.Sprintf("%v_%x", id.Tag, id.RequestHash)
}
