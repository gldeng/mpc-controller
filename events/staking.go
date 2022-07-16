package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning staking

type StakingTaskDoneEvent struct {
	TaskID common.Hash // TxHash

	ExportTxID       ids.ID
	ImportTxID       ids.ID
	AddDelegatorTxID ids.ID

	RequestID   uint64
	DelegateAmt uint64
	StartTime   uint64
	EndTime     uint64
	NodeID      ids.NodeID

	PubKeyHex     string
	CChainAddress common.Address
	PChainAddress ids.ShortID
	Nonce         uint64

	ParticipantPubKeys []string
}
