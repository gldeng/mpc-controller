package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ethereum/go-ethereum/common"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning rewarding

type StakingPeriodEndedEvent struct {
	AddDelegatorTxID ids.ID `copier:"must"`

	RequestID   uint64     `copier:"must"`
	DelegateAmt uint64     `copier:"must"`
	StartTime   uint64     `copier:"must"`
	EndTime     uint64     `copier:"must"`
	NodeID      ids.NodeID `copier:"must"`

	PubKeyHex     string      `copier:"must"`
	PChainAddress ids.ShortID `copier:"must"`
}

type RewardUTXOsFetchedEvent struct {
	AddDelegatorTxID ids.ID
	RewardUTXOs      []*avax.UTXO
}

type RewardUTXOsReportedEvent struct {
	AddDelegatorTxID ids.ID
	RewardUTXOIDs    []string
	TxHash           *common.Hash
}

type RewardingTaskStartedEvent struct {
	PChainAddress ids.ShortID
	UTXO          *avax.UTXO
}

type RewardingTaskDoneEvent struct {
}
