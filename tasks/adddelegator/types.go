package addDelegator

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
)

const (
	StatusInit Status = iota
	StatusSignReqSent
	StatusSignedTxReady
	StatusTxSent
	StatusDone
)

type Status int

const (
	SigLength = 65
)

type StakeParam struct {
	NodeID    ids.NodeID
	StartTime uint64
	EndTime   uint64
	UTXOs     []*avax.UTXO
}
