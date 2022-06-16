package hash256

import (
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func FromHex(hex string) common.Hash {
	return FromBytes(bytes.HexToBytes(hex))
}

func FromBytes(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}
