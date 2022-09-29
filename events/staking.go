package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common"
)

type StakeTaskDone struct {
	ReqNo   uint64
	Nonce   uint64
	ReqHash string

	DelegateAmt uint64
	StartTime   uint64
	EndTime     uint64
	NodeID      ids.NodeID

	AddDelegatorTxID ids.ID
	ExportTxID       ids.ID
	ImportTxID       ids.ID

	PubKeyHex     string
	CChainAddress common.Address
	PChainAddress ids.ShortID

	ParticipantPubKeys []string
}
