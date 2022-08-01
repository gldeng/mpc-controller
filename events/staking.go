package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
)

// ---------------------------------------------------------------------------------------------------------------------
// Events concerning staking

type StakeTaskDoneEvent struct {
	RequestID uint64
	Nonce     uint64
	TaskID    common.Hash // TxHash

	AddDelegatorTxID ids.ID
	ExportTxID       ids.ID
	ImportTxID       ids.ID

	DelegateAmt uint64
	StartTime   uint64
	EndTime     uint64
	NodeID      ids.NodeID

	PubKeyHex     string
	CChainAddress common.Address
	PChainAddress ids.ShortID

	ParticipantPubKeys []string
}
