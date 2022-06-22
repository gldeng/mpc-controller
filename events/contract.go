package events

import "math/big"

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning interact with contract

// MpcManager transactor

type ReportedGenPubKeyEvent struct {
	GroupIdHex       string
	Index            *big.Int
	GenPubKeyHex     string
	GenPubKeyHashHex string
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
