package bytes

import (
	"github.com/ethereum/go-ethereum/common"
)

func HexToBytes(hex string) []byte {
	return common.FromHex(hex)
}

func HexTo32Bytes(hex string) [32]byte {
	bytes := HexToBytes(hex)

	var res [32]byte
	copy(res[:], bytes)

	return res
}

func HexTo65Bytes(hex string) [65]byte {
	bytes := HexToBytes(hex)

	var res [65]byte
	copy(res[:], bytes)

	return res
}
