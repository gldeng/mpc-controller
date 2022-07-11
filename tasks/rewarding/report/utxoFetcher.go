package report

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"time"
)

const (
	// todo: to tune and make it long enough for consequential exporting reward process completed
	// to avoid double check on the same UTXO data set.
	checkUTXOInterval = time.Second * 120
)

// Accept event: *events.ReportedGenPubKeyEvent

// Emit event: *events.UTXOsFetchedEvent
// Emit event: *events.UTXOReportedEvent

type UTXOTracker struct {
	ContractAddr common.Address
	Logger       logger.Logger
	PChainClient platformvm.Client
	Publisher    dispatcher.Publisher
	Receipter    chain.Receipter
	Signer       *bind.TransactOpts
	Transactor   bind.ContractTransactor

	once               sync.Once
	lock               sync.Mutex
	genPubKeyEvtObjMap map[ids.ShortID]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *UTXOTracker) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.ReportedGenPubKeyEvent:
		eh.once.Do(func() {
			eh.genPubKeyEvtObjMap = make(map[ids.ShortID]*dispatcher.EventObject)
			go func() {
				eh.getAndReportUTXOs(ctx)
			}()
		})

		eh.lock.Lock()
		eh.genPubKeyEvtObjMap[evt.PChainAddress] = evtObj
		eh.lock.Unlock()
	}
}

func (eh *UTXOTracker) getAndReportUTXOs(ctx context.Context) {
	t := time.NewTicker(checkUTXOInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			eh.lock.Lock()
			for addr, evt := range eh.genPubKeyEvtObjMap {
				utxos, err := eh.getUTXOs(ctx, addr)
				if err != nil {
					eh.Logger.Error("Failed to get native UTXOs", []logger.Field{{"error", err}}...)
					continue
				}

				if len(utxos) == 0 {
					eh.Logger.Debug("Found no native UTXOs", []logger.Field{{"pChainAddress", addr}}...)
					continue
				}

				mpcUTXOs := myAvax.MpcUTXOsFromUTXOs(utxos)

				utxoFetchedEvt := &events.UTXOsFetchedEvent{
					NativeUTXOs: utxos,
					MpcUTXOs:    mpcUTXOs,
				}

				copier.Copy(&utxoFetchedEvt, evt.Event.(*events.ReportedGenPubKeyEvent))

				utxoFetchedEvtObj := dispatcher.NewRootEventObject("UTXOTracker", utxoFetchedEvt, ctx)
				eh.Publisher.Publish(ctx, utxoFetchedEvtObj)

				groupIdBytes := bytes.HexTo32Bytes(utxoFetchedEvt.GroupIdHex)
				partiIndex := utxoFetchedEvt.PartiIndex
				genPubKeyBytes := bytes.HexToBytes(utxoFetchedEvt.GenPubKeyHex)
				for _, utxo := range utxos {
					txHash, err := eh.reportUTXO(ctx, groupIdBytes, partiIndex, genPubKeyBytes, utxo.TxID, utxo.OutputIndex)
					if err != nil {
						eh.Logger.Error("Failed to report UTXO", []logger.Field{{"error", err}}...)
						continue
					}

					utxoReportedEvt := &events.UTXOReportedEvent{
						TxID:           utxo.TxID,
						OutputIndex:    utxo.OutputIndex,
						TxHash:         txHash,
						GenPubKeyBytes: genPubKeyBytes,
						GroupIDBytes:   groupIdBytes,
						PartiIndex:     partiIndex,
					}
					eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(utxoFetchedEvtObj, "UTXOTracker", utxoReportedEvt, ctx))
				}
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (eh *UTXOTracker) getUTXOs(ctx context.Context, addr ids.ShortID) ([]*avax.UTXO, error) {
	var utxoBytesArr [][]byte
	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		var err error
		utxoBytesArr, _, _, err = eh.PChainClient.GetUTXOs(ctx, []ids.ShortID{addr}, 0, addr, ids.ID{})
		if err != nil {
			return errors.Wrap(err, "failed to request native UTXOs")
		}
		return nil
	})

	utxos, err := myAvax.ParseUTXOs(utxoBytesArr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse native UTXO bytes")
	}

	var utxosFiltered []*avax.UTXO
	for _, utxo := range utxos {
		if utxo != nil {
			utxosFiltered = append(utxosFiltered, utxo)
		}
	}
	return utxosFiltered, nil
}

func (eh *UTXOTracker) reportUTXO(ctx context.Context, groupId [32]byte, partiIndex *big.Int, genPubKey []byte, txID [32]byte, outputIndex uint32) (txHash *common.Hash, err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor) // todo: extract to reuse in multi flows.
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var tx *types.Transaction

	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		tx, err = transactor.ReportUTXO(eh.Signer, groupId, partiIndex, genPubKey, txID, outputIndex)
		if err != nil {
			err = errors.Wrap(err, "failed to ReportUTXO")
			return err
		}

		time.Sleep(time.Second * 3)

		var rcpt *types.Receipt
		rcpt, err = eh.Receipter.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return errors.WithStack(err)
		}

		if rcpt.Status != 1 {
			err = errors.Errorf("failed to report UTXO")
			return err
		}

		newTxHash := tx.Hash()
		txHash = &newTxHash
		return nil
	})
	return
}
