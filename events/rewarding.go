package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning rewarding

type StakingPeriodEndedEvent struct {
	AddDelegatorTxID ids.ID

	RequestID   uint64
	DelegateAmt uint64
	StartTime   uint64
	EndTime     uint64
	NodeID      ids.NodeID

	PubKeyHex     string
	PChainAddress ids.ShortID
}

type RewardingTaskStartedEvent struct {
	PChainAddress ids.ShortID
	UTXO          *avax.UTXO
}

type RewardingTaskDoneEvent struct {
}
