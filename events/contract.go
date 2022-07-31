package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning interact with contract

// MpcManager transactor

type ReportedGenPubKeyEvent struct {
	GroupIdHex       string
	MyPartiIndex     *big.Int
	GenPubKeyHex     string
	GenPubKeyHashHex string
	CChainAddress    common.Address
	PChainAddress    ids.ShortID
}

type JoinRequestEvent struct {
	RequestId  *big.Int
	PartiIndex *big.Int
}

type JoinedRequestEvent struct {
	TxHashHex  string
	RequestId  *big.Int
	PartiIndex *big.Int
}
