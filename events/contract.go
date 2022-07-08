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
	Index            *big.Int
	GenPubKeyHex     string
	GenPubKeyHashHex string
	CChainAddress    common.Address
	PChainAddress    ids.ShortID
}

type JoinRequestEvent struct {
	RequestId *big.Int
	Index     *big.Int
}

type JoinedRequestEvent struct {
	TxHashHex string
	RequestId *big.Int
	Index     *big.Int
}
