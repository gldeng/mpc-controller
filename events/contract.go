package events

import "math/big"

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning interact with contract

// MpcManager transactor

type RequestJoinRequestEvent struct {
	RequestId *big.Int
	Index     *big.Int
}
