package bytes

import (
	"github.com/ethereum/go-ethereum/common"
)

func BytesToHexArr(bytesArr [][]byte) []string {
	var res []string
	for _, bytes := range bytesArr {
		res = append(res, BytesToHex(bytes))
	}
	return res
}

func Bytes32ToHexArr(bytesArr [][32]byte) []string {
	var res []string
	for _, bytes := range bytesArr {
		res = append(res, Bytes32ToHex(bytes))
	}
	return res
}

func BytesToHex(bytes []byte) string {
	return common.Bytes2Hex(bytes)
}

func Bytes32ToHex(bytes [32]byte) string {
	return BytesToHex(bytes[:])
}
