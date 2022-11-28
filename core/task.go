package core

import (
	"github.com/ava-labs/avalanchego/ids"
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
	RequestConfirmations(opts *bind.CallOpts, groupId [32]byte, requestHash [32]byte) (*big.Int, error)
}

type TaskContext interface {
	MpcManager
	GetLogger() logger.Logger
	GetNetwork() *NetworkContext
	GetMpcClient() MpcClient
	IssueCChainTx(txBytes []byte) (ids.ID, error)
	IssuePChainTx(txBytes []byte) (ids.ID, error)
	CheckEthTx(txHash common.Hash) (TxStatus, error)
	CheckCChainTx(id ids.ID) (TxStatus, error)
	CheckPChainTx(id ids.ID) (TxStatus, error)
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
