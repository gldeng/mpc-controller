package mpc_controller

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/storage"
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
	JoinRequest(requestId *big.Int, myIndex *big.Int) (*types.Transaction, error)
}

type TransactorReportGeneratedKey interface {
	ReportGeneratedKey(groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error)
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

type StorerStoreGroupInfo interface {
	StoreGroupInfo(g *storage.GroupInfo) error
}

type StorerLoadGroupInfo interface {
	LoadGroupInfo(groupIdHex string) (*storage.GroupInfo, error)
}

type StorerLoadGroupInfos interface {
	LoadGroupInfos() ([]*storage.GroupInfo, error)
}

type StorerStoreParticipantInfo interface {
	StoreParticipantInfo(p *storage.ParticipantInfo) error
}

type StorerLoadParticipantInfo interface {
	LoadParticipantInfo(pubKeyHashHex, groupId string) (*storage.ParticipantInfo, error)
}

type StorerLoadParticipantInfos interface {
	LoadParticipantInfos(pubKeyHashHex string) ([]*storage.ParticipantInfo, error)
}

type StorerStoreGeneratedPubKeyInfo interface {
	StoreGeneratedPubKeyInfo(genPubKeyInfo *storage.GeneratedPubKeyInfo) error
}

type StorerLoadGeneratedPubKeyInfo interface {
	LoadGeneratedPubKeyInfo(pubKeyHashHex string) (*storage.GeneratedPubKeyInfo, error)
}

type StorerLoadGeneratedPubKeyInfos interface {
	LoadGeneratedPubKeyInfos(groupIdHexs []string) ([]*storage.GeneratedPubKeyInfo, error)
}

type StorerStoreKeygenRequestInfo interface {
	StoreKeygenRequestInfo(keygenReqInfo *storage.KeygenRequestInfo) error
}

type StorerLoadKeygenRequestInfo interface {
	LoadKeygenRequestInfo(reqIdHex string) (*storage.KeygenRequestInfo, error)
}
