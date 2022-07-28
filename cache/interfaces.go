package cache

import (
	"github.com/avalido/mpc-controller/events"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type ICache interface {
	MyIndexGetter
	GeneratedPubKeyInfoGetter
	ParticipantKeysGetter
	NormalizedParticipantKeysGetter
	IsParticipantChecker
}

type MyIndexGetter interface {
	GetMyIndex(myPubKeyHashHex, genPubKeyHashHex string) *big.Int
}

type GeneratedPubKeyInfoGetter interface {
	GetGeneratedPubKeyInfo(genPubKeyHashHex string) *events.GeneratedPubKeyInfo
}

type ParticipantKeysGetter interface {
	GetParticipantKeys(genPubKeyHash common.Hash, indices []*big.Int) []string
}

type NormalizedParticipantKeysGetter interface {
	GetNormalizedParticipantKeys(genPubKeyHash common.Hash, indices []*big.Int) ([]string, error)
}

type IsParticipantChecker interface {
	IsParticipant(myPubKeyHash string, genPubKeyHash string, participantIndices []*big.Int) bool
}

// ---------

// implemented by *standard sync.Map, which support concurrent manipulation.

type SyncMapCache interface {
	Delete(key any)
	Load(key any) (value any, ok bool)
	LoadAndDelete(key any) (value any, loaded bool)
	LoadOrStore(key, value any) (actual any, loaded bool)
	Range(f func(key, value any) bool)
	Store(key, value any)
}
