package events

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ethereum/go-ethereum/common"
)

type StakeAtomicTaskHandled struct {
	StakeTaskBasic

	ExportTxID ids.ID
	ImportTxID ids.ID

	UTXOsToStake []*avax.UTXO `json:"-"`
}

type StakeAddDelegatorTaskDone struct {
	StakeTaskBasic

	AddDelegatorTxID ids.ID
}

type StakeTaskBasic struct {
	ReqNo   uint64
	Nonce   uint64
	ReqHash string

	DelegateAmt uint64
	StartTime   uint64
	EndTime     uint64
	NodeID      ids.NodeID

	StakePubKey   string // compressed hex string
	CChainAddress common.Address
	PChainAddress ids.ShortID

	JoinedPubKeys []string // compressed hex string
}
