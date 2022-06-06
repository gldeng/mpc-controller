package mpc_controller

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	GetGroup(groupId [32]byte) (Participants [][]byte, Threshold *big.Int, err error)
}

// Transactor

type TransactorJoinRequest interface {
	JoinRequest(opts *bind.TransactOpts, requestId *big.Int, myIndex *big.Int) (*types.Transaction, error)
}

type TransactorReportGeneratedKey interface {
	ReportGeneratedKey(opts *bind.TransactOpts, groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error)
}

// Filters

type WatcherParticipantAdded interface {
	WatchParticipantAdded(publicKey [][]byte) (<-chan *contract.MpcManagerParticipantAdded, error)
}

type WatcherKeygenRequestAdded interface {
	WatchKeygenRequestAdded(groupId [][32]byte) (<-chan *contract.MpcManagerKeygenRequestAdded, error)
}

type WatcherKeyGenerated interface {
	WatchKeyGenerated(groupId [][32]byte) (<-chan *contract.MpcManagerKeyGenerated, error)
}

type WatcherStakeRequestAdded interface {
	WatchStakeRequestAdded(publicKey [][]byte) (<-chan *contract.MpcManagerStakeRequestAdded, error)
}

type WatcherStakeRequestStarted interface {
	WatchStakeRequestStarted(publicKey [][]byte) (<-chan *contract.MpcManagerStakeRequestStarted, error)
}

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding storage

// Group

type StorerStoreGroupInfo interface {
	StoreGroupInfo(g *storage.GroupInfo) error
}

type StorerLoadGroupInfo interface {
	LoadGroupInfo(groupIdHex string) (*storage.GroupInfo, error)
}

type StorerLoadGroupInfos interface {
	LoadGroupInfos() ([]*storage.GroupInfo, error)
}

// Participant

type StorerStoreParticipantInfo interface {
	StoreParticipantInfo(p *storage.ParticipantInfo) error
}

type StorerLoadParticipantInfo interface {
	LoadParticipantInfo(pubKeyHashHex, groupId string) (*storage.ParticipantInfo, error)
}

type StorerLoadParticipantInfos interface {
	LoadParticipantInfos(pubKeyHashHex string) ([]*storage.ParticipantInfo, error)
}

type StorerGetParticipantIndex interface {
	GetIndex(partiPubKeyHashHex, genPubKeyHexHex string) (*big.Int, error)
}

type StorerGetGroupIds interface {
	GetGroupIds(partiPubKeyHashHex string) ([][32]byte, error)
}

type StorerGetPubKeys interface {
	GetPubKeys(partiPubKeyHashHex string) ([][]byte, error)
}

// Generated public key

type StorerStoreGeneratedPubKeyInfo interface {
	StoreGeneratedPubKeyInfo(genPubKeyInfo *storage.GeneratedPubKeyInfo) error
}

type StorerLoadGeneratedPubKeyInfo interface {
	LoadGeneratedPubKeyInfo(pubKeyHashHex string) (*storage.GeneratedPubKeyInfo, error)
}

type StorerLoadGeneratedPubKeyInfos interface {
	LoadGeneratedPubKeyInfos(groupIdHexs []string) ([]*storage.GeneratedPubKeyInfo, error)
}

type StorerGetPariticipantKeys interface {
	GetPariticipantKeys(genPubKeyHashHex string, indices []*big.Int) ([]string, error)
}

// Keygen request info

type StorerStoreKeygenRequestInfo interface {
	StoreKeygenRequestInfo(keygenReqInfo *storage.KeygenRequestInfo) error
}

type StorerLoadKeygenRequestInfo interface {
	LoadKeygenRequestInfo(reqIdHex string) (*storage.KeygenRequestInfo, error)
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
