package events

import "github.com/ava-labs/avalanchego/ids"

const (
	TxKindCChainExport TxKind = iota
	TxKindCChainImport

	TxKindPChainExport
	TxKindPChainImport
	TxKindPChainAddDelegator
)

type TxKind int
type Chain int

type TxCommitted struct { // P-Chain
	ReqID string
	Kind  TxKind
	TxID  ids.ID
}

type TxAccepted struct { // C-Chain
	ReqID string
	Kind  TxKind
	TxID  ids.ID
}
