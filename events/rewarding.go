package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
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
	PubKeyHex        string
}

type RewardedStakeReportedEvent struct {
	AddDelegatorTxID ids.ID
	PubKeyHex        string
	GroupIDHex       string
	MyIndex          *big.Int
	TxHash           *common.Hash
}

type ExportRewardRequestAddedEvent struct {
	AddDelegatorTxID ids.ID
	PublicKeyHash    common.Hash
	TxHash           common.Hash
}

type JoinedExportRewardRequestEvent struct {
	GroupIDHex       string
	MyIndex          *big.Int
	PubKeyHex        string
	AddDelegatorTxID ids.ID
	TxHash           common.Hash
}

type ExportRewardRequestStartedEvent struct {
	AddDelegatorTxID   ids.ID
	PublicKeyHash      common.Hash
	ParticipantIndices []*big.Int
	TxHash             common.Hash
}

type RewardExportedEvent struct {
	AddDelegatorTxID ids.ID
	ExportedTxID     ids.ID
	ImportedTxID     ids.ID
	// todo: more fields
}
