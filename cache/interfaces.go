package cache

import (
	"github.com/avalido/mpc-controller/events"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type MyIndexGetter interface {
	GetMyIndex(myPubKeyHashHex, genPubKeyHashHex string) *big.Int
}

type GeneratedPubKeyInfoGetter interface {
	GetGeneratedPubKeyInfo(genPubKeyHashHex string) *events.GeneratedPubKeyInfo
}

type ParticipantKeysGetter interface {
	GetParticipantKeys(genPubKeyHash common.Hash, indices []*big.Int) []string
}
