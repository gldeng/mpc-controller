package cache

import (
	"github.com/avalido/mpc-controller/events"
	"math/big"
)

type MyIndexGetter interface {
	GetMyIndex(myPubKeyHashHex, genPubKeyHashHex string) *big.Int
}

type GeneratedPubKeyInfoGetter interface {
	GetGeneratedPubKeyInfo(genPubKeyHashHex string) *events.GeneratedPubKeyInfo
}
