package bytes

import (
	"github.com/ethereum/go-ethereum/common"
)

func HexToBytesArr(hexes []string) [][]byte {
	var res [][]byte
	for _, hex := range hexes {
		res = append(res, HexToBytes(hex))
	}
	return res
}

func HexTo32BytesArr(hexes []string) [][32]byte {
	var res [][32]byte
	for _, hex := range hexes {
		res = append(res, HexTo32Bytes(hex))
	}
	return res
}

func HexToBytes(hex string) []byte {
	return common.FromHex(hex)
}

func HexTo32Bytes(hex string) [32]byte {
	bytes := HexToBytes(hex)

	var res [32]byte
	copy(res[:], bytes)

	return res
}
