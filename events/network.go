package events

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning network

type ContractFiltererReconnectedEvent struct {
	Filterer bind.ContractFilterer
}
