package events

import "github.com/ava-labs/avalanchego/vms/components/avax"

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning rewarding

type RewardingTaskStartedEvent struct {
	UTXO *avax.UTXO
}

type RewardingTaskDoneEvent struct {
}
