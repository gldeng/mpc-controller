package cache

import "math/big"

type MyIndexGetter interface {
	GetMyIndex(myPubKeyHashHex, genPubKeyHashHex string) *big.Int
}
