package core

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Task interface {
	GetId() string
	Next(ctx TaskContext) ([]Task, error)
	IsDone() bool
	FailedPermanently() bool
	RequiresNonce() bool
}

type MpcManager interface {
	GetGroup(opts *bind.CallOpts, groupId [32]byte) ([][]byte, error)
	ReportGeneratedKey(opts *bind.TransactOpts, participantId [32]byte, generatedPublicKey []byte) (*common.Hash, error)
}

type TaskContext interface {
	MpcManager
	GetLogger() logger.Logger
	GetNetwork() *chain.NetworkContext
	GetMpcClient() MpcClient
	IssueCChainTx(txBytes []byte) (ids.ID, error)
	IssuePChainTx(txBytes []byte) (ids.ID, error)
	CheckEthTx(txHash common.Hash) (TxStatus, error)
	CheckCChainTx(id ids.ID) (TxStatus, error)
	CheckPChainTx(id ids.ID) (TxStatus, error)
	NonceAt(account common.Address) (uint64, error)
	Emit(event interface{})
	GetDb() storage.SlimDb
	GetEventID(event string) (common.Hash, error)
	GetMyPublicKey() ([]byte, error)
	LoadGroup() (*types.Group, error)
	GetParticipantID() storage.ParticipantId
}

type TaskContextFactory = func() TaskContext
type TaskSubmitter interface {
	Submit(task Task) error
}
