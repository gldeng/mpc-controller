package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ethereum/go-ethereum/common"
)

type StakeAtomicTaskDone struct {
	ReqNo   uint64
	Nonce   uint64
	ReqHash string

	DelegateAmt uint64
	StartTime   uint64
	EndTime     uint64
	NodeID      ids.NodeID

	ExportTxID ids.ID
	ImportTxID ids.ID

	UTXOsToStake []*avax.UTXO

	PubKeyHex     string
	CChainAddress common.Address
	PChainAddress ids.ShortID

	ParticipantPubKeys []string
}

type StakeAddDelegatorTaskDone struct {
	ReqNo   uint64
	Nonce   uint64
	ReqHash string

	DelegateAmt uint64
	StartTime   uint64
	EndTime     uint64
	NodeID      ids.NodeID

	AddDelegatorTxID ids.ID

	PubKeyHex     string
	CChainAddress common.Address
	PChainAddress ids.ShortID

	ParticipantPubKeys []string
}
