package bytes

import (
	"github.com/ethereum/go-ethereum/common"
)

func BytesToHex(bytes []byte) string {
	return common.Bytes2Hex(bytes)
}

func Bytes32ToHex(bytes [32]byte) string {
	return BytesToHex(bytes[:])
}

func Bytes65ToHex(bytes [65]byte) string {
	return BytesToHex(bytes[:])
}
