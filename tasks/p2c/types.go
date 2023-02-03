package p2c

import (
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
)

type ImportedEvent struct {
	Tx *txs.ImportTx
}

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
