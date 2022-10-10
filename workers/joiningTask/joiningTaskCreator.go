package joiningTask

import (
	"context"
	"fmt"
	"github.com/alitto/pond"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/sequencer"
	"github.com/avalido/mpc-controller/workers"
	"github.com/ethereum/go-ethereum/common"
	kbcevents "github.com/kubecost/events"
	"time"
)

type JoiningTaskCreator struct {
	EthClient chain.EthClient
	TxIssuer  txissuer.TxIssuer
	Logger    logger.Logger

	signedTxChan chan any

	sequencers map[common.Address]sequencer.Sequencer

	workerPool *pond.WorkerPool
	dispatcher kbcevents.Dispatcher[*events.TxApproved]
}

func (c *JoiningTaskCreator) Run(ctx context.Context) {
	//txApprovedEvtHandler := func(evt events.TxApproved) {
	//	fmt.Println(event.Message)
	//}
	//
	//txApprovedEvtFilter := func(evt events.TxApproved) bool {
	//	return evt.Kind == events.TxKindCChainExport || evt.Kind == events.TxKindCChainImport
	//}
	//c.AddFilteredEventHandler(txApprovedEvtHandler, txApprovedEvtFilter)
	go c.IssueSignedTx(ctx)
}

func (c *JoiningTaskCreator) AddSignedTx(tx any) { // accept SignedTxWithNonce and SignedTx
	c.signedTxChan <- tx
}

func (c *JoiningTaskCreator) IssueSignedTx(ctx context.Context) {
	issueT := time.NewTicker(time.Second)
	defer issueT.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-c.signedTxChan:
			// Add SignedTxWithNonce to Sequencer
			txWithNonceVal, ok := tx.(workers.SignedTxWithNonce)
			if ok {
				seq, ok := c.sequencers[txWithNonceVal.Address()]
				if !ok {
					c.sequencers[txWithNonceVal.Address()] = &sequencer.AscendingSequencer{}
					seq = c.sequencers[txWithNonceVal.Address()]
				}

				seq.AddThenSort(txWithNonceVal)
				continue
			}

			// Submit SignedTx to worker pool
			txVal, ok := tx.(workers.SignedTx)
			if ok {
				c.workerPool.Submit(func() {
					// Issue signed ExportTx parallelly
					tx := txissuer.Tx{
						ReqID: txVal.ReqID(),
						Kind:  txVal.Kind(),
						Bytes: txVal.SignedBytes(),
					}
					err := c.TxIssuer.IssueTx(ctx, &tx)
					if err != nil {
						c.Logger.ErrorOnError(err, "Failed to issue SignedTx", []logger.Field{{"failedToIssueSignedTx",
							fmt.Sprintf("reqID:%v kind:%v", tx.ReqID, tx.Kind)}}...)
					}
				})
			}

		case <-issueT.C:
			for addr, seq := range c.sequencers {
				// Sync address nonce
				if seq.IsEmpty() {
					continue
				}
				addressNonce, err := c.EthClient.NonceAt(ctx, addr, nil)
				if err != nil {
					c.Logger.ErrorOnError(err, fmt.Sprintf("Failed to query nonce for address %v", addr))
					continue
				}

				objs := seq.ContinuousObjs(addressNonce)
				if len(objs) == 0 {
					continue
				}
				for _, obj := range objs {
					// Issue signed ExportTx sequentially
					txVal := obj.(workers.SignedTxWithNonce)
					tx := txissuer.Tx{
						ReqID: txVal.ReqID(),
						Kind:  txVal.Kind(),
						Bytes: txVal.SignedBytes(),
					}
					err = c.TxIssuer.IssueTx(ctx, &tx)
					if err != nil {
						c.Logger.ErrorOnError(err, "Failed to issue signedTxWithNonce", []logger.Field{{"failedToIssueSignedTxWithNonce",
							fmt.Sprintf("reqID:%v kind:%v address:%v nonce:%v", tx.ReqID, tx.Kind, addr, txVal.Nonce())}}...)
						continue
					}
				}

				seq.TrimLeft(objs[len(objs)-1].Nonce())
			}
		}
	}
}

// todo: improvement to avoid absolute resource leak

func (c *JoiningTaskCreator) Stop() {
	c.Stop()
}
