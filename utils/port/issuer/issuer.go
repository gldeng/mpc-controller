package issuer

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/pkg/errors"
)

const (
	INVALID IssueOrder = "Invalid"
	P2C     IssueOrder = "P2C"
	C2P     IssueOrder = "C2P"
)

type IssueOrder string

//var _ porter.TxIssuer = (*Issuer)(nil)

// todo: consider back off retry strategy on network partition

type Issuer struct {
	CChainIssueClient chain.CChainIssuer
	PChainIssueClient chain.PChainIssuer
	IssueOrder        IssueOrder
}

func (i *Issuer) IssueExportTx(ctx context.Context, exportTxBytes []byte) (id ids.ID, order IssueOrder, err error) {
	switch i.IssueOrder {
	case P2C:
		order = P2C
		id, err = i.PChainIssueClient.IssueTx(ctx, exportTxBytes)
		return
	case C2P:
		order = C2P
		id, err = i.CChainIssueClient.IssueTx(ctx, exportTxBytes)
		return
	default:
		return ids.ID{}, INVALID, errors.New("invalid order")
	}
}

func (i *Issuer) IssueImportTx(ctx context.Context, importTxBytes []byte) (id ids.ID, order IssueOrder, err error) {
	switch i.IssueOrder {
	case P2C:
		order = P2C
		id, err = i.CChainIssueClient.IssueTx(ctx, importTxBytes)
		return
	case C2P:
		order = C2P
		id, err = i.PChainIssueClient.IssueTx(ctx, importTxBytes)
		return
	default:
		return ids.ID{}, INVALID, errors.New("invalid order")
	}
}
