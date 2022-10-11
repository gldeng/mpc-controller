package txissuer

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/status"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	StatusIssued Status = iota
	StatusApproved
	StatusFailed
)

// todo: modify implementation for new architecture.

type TxIssuer interface {
	IssueTx(ctx context.Context, tx *Tx) error
	TrackTx(ctx context.Context, tx *Tx) (Status, error)
}

type Status int

type Tx struct {
	ReqID string
	Kind  events.TxKind
	Bytes []byte
	txID  ids.ID
}

// todo: add retry mechanism
// todo: close dispatchers when they no longer needed.

type MyTxIssuer struct {
	Logger       logger.Logger
	CChainClient evm.Client
	PChainClient platformvm.Client

	dispatcher kbcevents.Dispatcher[*events.TxApproved]

	pendingTx *sync.Map
	once      *sync.Once
}

func (t *MyTxIssuer) IssueTx(ctx context.Context, tx *Tx) error {
	t.once.Do(func() {
		go t.trackTx(ctx)
	})
	var txID ids.ID
	var err error
	switch tx.Kind {
	case events.TxKindCChainExport, events.TxKindCChainImport:
		txID, err = t.CChainClient.IssueTx(ctx, tx.Bytes)
		if err != nil {
			return errors.Wrapf(err, "failed to issue C-Chain tx")
		}
		t.Logger.Info("Issued C-Chain tx", []logger.Field{{"issuedTx", tx}}...)
	case events.TxKindPChainExport, events.TxKindPChainImport, events.TxKindPChainAddDelegator:
		txID, err = t.PChainClient.IssueTx(ctx, tx.Bytes)
		if err != nil {
			return errors.Wrapf(err, "failed to issue P-Chain tx")
		}
		t.Logger.Info("Issued P-Chain tx", []logger.Field{{"issuedTx", tx}}...)
	}
	tx.txID = txID
	t.pendingTx.Store(txID, tx)
	return nil
}

func (t *MyTxIssuer) trackTx(ctx context.Context) {
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			t.pendingTx.Range(func(key, value any) bool {
				tx := value.(*Tx)
				switch tx.Kind {
				case events.TxKindCChainExport, events.TxKindCChainImport:
					status, err := t.CChainClient.GetAtomicTxStatus(ctx, tx.txID)
					if err != nil {
						t.Logger.ErrorOnError(err, "Failed to check C-Chain tx")
						break
					}
					switch status {
					case evm.Dropped:
						t.Logger.Warn("C-Chain tx dropped", []logger.Field{{"droppedTx", tx}}...)
						t.pendingTx.Delete(tx.txID)
					case evm.Processing:
						t.Logger.Debug("C-Chain tx processing", []logger.Field{{"processingTx", tx}}...)
					case evm.Accepted:
						txAcc := events.TxApproved{
							ReqID: tx.ReqID,
							Kind:  tx.Kind,
							TxID:  tx.txID,
						}
						t.dispatcher.Dispatch(&txAcc)
						t.Logger.Info("C-Chain tx accepted", []logger.Field{{"acceptedTx", tx}}...)
						t.pendingTx.Delete(tx.txID)
					}
				case events.TxKindPChainExport, events.TxKindPChainImport, events.TxKindPChainAddDelegator:
					statusResp, err := t.PChainClient.GetTxStatus(ctx, tx.txID)
					if err != nil {
						t.Logger.ErrorOnError(err, "Failed to check P-Chain tx")
						break
					}
					switch statusResp.Status {
					case status.Aborted:
						t.Logger.Warn("P-Chain tx aborted", []logger.Field{{"abortedTx", tx}}...)
						t.pendingTx.Delete(tx.txID)
					case status.Processing:
						t.Logger.Debug("P-Chain tx processing", []logger.Field{{"processingTx", tx}}...)
					case status.Dropped:
						t.Logger.Warn("P-Chain tx dropped", []logger.Field{{"droppedTx", tx}}...)
						t.pendingTx.Delete(tx.txID)
					case status.Committed:
						txCmt := events.TxApproved{
							ReqID: tx.ReqID,
							Kind:  tx.Kind,
							TxID:  tx.txID,
						}
						t.dispatcher.Dispatch(&txCmt)
						t.Logger.Info("Tx committed", []logger.Field{{"acceptedTx", tx}}...)
						t.pendingTx.Delete(tx.txID)
					}
				}
				return true
			})
		}
	}
}
