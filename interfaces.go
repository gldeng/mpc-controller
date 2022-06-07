package mpc_controller

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding mpc-controller service

type MpcControllerService interface {
	Start(ctx context.Context) error
}

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding MpcManager contract

// Callers

type CallerGetGroup interface {
	GetGroup(ctx context.Context, groupId [32]byte) (Participants [][]byte, Threshold *big.Int, err error)
}

// Transactor

type TransactorJoinRequest interface {
	JoinRequest(ctx context.Context, opts *bind.TransactOpts, requestId *big.Int, myIndex *big.Int) (*types.Transaction, error)
}

type TransactorReportGeneratedKey interface {
	ReportGeneratedKey(ctx context.Context, opts *bind.TransactOpts, groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error)
}

// Filters

type WatcherParticipantAdded interface {
	WatchParticipantAdded(ctx context.Context, publicKey [][]byte) (<-chan *contract.MpcManagerParticipantAdded, error)
}

type WatcherKeygenRequestAdded interface {
	WatchKeygenRequestAdded(ctx context.Context, groupId [][32]byte) (<-chan *contract.MpcManagerKeygenRequestAdded, error)
}

type WatcherKeyGenerated interface {
	WatchKeyGenerated(ctx context.Context, groupId [][32]byte) (<-chan *contract.MpcManagerKeyGenerated, error)
}

type WatcherStakeRequestAdded interface {
	WatchStakeRequestAdded(ctx context.Context, publicKey [][]byte) (<-chan *contract.MpcManagerStakeRequestAdded, error)
}

type WatcherStakeRequestStarted interface {
	WatchStakeRequestStarted(ctx context.Context, publicKey [][]byte) (<-chan *contract.MpcManagerStakeRequestStarted, error)
}

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding storage

// Group

type StorerStoreGroupInfo interface {
	StoreGroupInfo(ctx context.Context, g *storage.GroupInfo) error
}

type StorerLoadGroupInfo interface {
	LoadGroupInfo(ctx context.Context, groupIdHex string) (*storage.GroupInfo, error)
}

type StorerLoadGroupInfos interface {
	LoadGroupInfos(ctx context.Context) ([]*storage.GroupInfo, error)
}

// Participant

type StorerStoreParticipantInfo interface {
	StoreParticipantInfo(ctx context.Context, p *storage.ParticipantInfo) error
}

type StorerLoadParticipantInfo interface {
	LoadParticipantInfo(ctx context.Context, pubKeyHashHex, groupId string) (*storage.ParticipantInfo, error)
}

type StorerLoadParticipantInfos interface {
	LoadParticipantInfos(ctx context.Context, pubKeyHashHex string) ([]*storage.ParticipantInfo, error)
}

type StorerGetParticipantIndex interface {
	GetIndex(ctx context.Context, partiPubKeyHashHex, genPubKeyHexHex string) (*big.Int, error)
}

type StorerGetGroupIds interface {
	GetGroupIds(ctx context.Context, partiPubKeyHashHex string) ([][32]byte, error)
}

type StorerGetPubKeys interface {
	GetPubKeys(ctx context.Context, partiPubKeyHashHex string) ([][]byte, error)
}

// Generated public key

type StorerStoreGeneratedPubKeyInfo interface {
	StoreGeneratedPubKeyInfo(ctx context.Context, genPubKeyInfo *storage.GeneratedPubKeyInfo) error
}

type StorerLoadGeneratedPubKeyInfo interface {
	LoadGeneratedPubKeyInfo(ctx context.Context, pubKeyHashHex string) (*storage.GeneratedPubKeyInfo, error)
}

type StorerLoadGeneratedPubKeyInfos interface {
	LoadGeneratedPubKeyInfos(ctx context.Context, groupIdHexs []string) ([]*storage.GeneratedPubKeyInfo, error)
}

type StorerGetPariticipantKeys interface {
	GetPariticipantKeys(ctx context.Context, genPubKeyHashHex string, indices []*big.Int) ([]string, error)
}

// Keygen request info

type StorerStoreKeygenRequestInfo interface {
	StoreKeygenRequestInfo(ctx context.Context, keygenReqInfo *storage.KeygenRequestInfo) error
}

type StorerLoadKeygenRequestInfo interface {
	LoadKeygenRequestInfo(ctx context.Context, reqIdHex string) (*storage.KeygenRequestInfo, error)
}

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding mpc-client

type MpcClientKeygen interface {
	Keygen(ctx context.Context, keygenReq *core.KeygenRequest) error
}

type MpcClientSign interface {
	Sign(ctx context.Context, signReq *core.SignRequest) error
}

type MpcClientResult interface {
	Result(ctx context.Context, reqID string) (*core.Result, error)
}

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding eth client

type EthClientTransactionReceipt interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type EthClientNonceAt interface {
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
}

type EthClientBalanceAt interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}
