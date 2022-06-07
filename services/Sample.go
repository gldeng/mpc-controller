package services

import (
	"github.com/ethereum/go-ethereum/common"
	"math/rand"
	"time"
)

func Sample(arr []common.Hash) []common.Hash {
	var out []common.Hash
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	for _, txHash := range arr {
		if r.Intn(1) == 0 {
			out = append(out, txHash)
		}
	}
	return out
}
