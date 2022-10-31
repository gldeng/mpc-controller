package addDelegator

import "github.com/ava-labs/avalanchego/ids"

const (
	StatusInit Status = iota
	StatusSignReqSent
	StatusTxSent
	StatusDone
)

type Status int

const (
	SigLength = 65
)

type Request struct {
	NodeID    ids.NodeID
	StartTime uint64
	EndTime   uint64
}
