package issuer

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/utils/port/porter"
	"github.com/pkg/errors"
)

const (
	P2C IssueOrder = iota
	C2P
)

type IssueOrder int

var _ porter.TxIssuer = (*Issuer)(nil)

// todo: consider back off retry strategy on network partition

type Issuer struct {
	CChainIssueClient chain.Issuer
	PChainIssueClient chain.Issuer
	IssueOrder        IssueOrder
}

func (i *Issuer) IssueExportTx(ctx context.Context, exportTxBytes []byte) (ids.ID, error) {
	switch i.IssueOrder {
	case P2C:
		return i.PChainIssueClient.IssueTx(ctx, exportTxBytes)
	case C2P:
		return i.CChainIssueClient.IssueTx(ctx, exportTxBytes)
	default:
		return ids.ID{}, errors.New("invalid order")
	}
}

func (i *Issuer) IssueImportTx(ctx context.Context, importTxBytes []byte) (ids.ID, error) {
	switch i.IssueOrder {
	case P2C:
		return i.CChainIssueClient.IssueTx(ctx, importTxBytes)
	case C2P:
		return i.PChainIssueClient.IssueTx(ctx, importTxBytes)
	default:
		return ids.ID{}, errors.New("invalid order")
	}
}
