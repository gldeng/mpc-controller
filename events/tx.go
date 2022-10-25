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

type TxApproved struct {
	ReqID string
	Kind  TxKind
	TxID  ids.ID
}
