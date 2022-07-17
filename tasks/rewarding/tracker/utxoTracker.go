package tracker

import (
	"context"
	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
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
	"math/rand"
	"sync"
	"time"
)

const (
	// todo: to tune and make it long enough for consequential exporting reward process completed
	// to avoid double check on the same UTXO data set.
	checkUTXOInterval = time.Minute * 5
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
					ok, err := eh.checkRewardUTXO(ctx, utxo.TxID)
					if err != nil {
						eh.Logger.Error("Failed to check reward UTXO for txID", []logger.Field{
							{"txID", utxo.TxID},
							{"error", err}}...)
						continue
					}

					if !ok {
						ok, err := eh.checkImportTx(ctx, utxo.TxID)
						if err != nil {
							eh.Logger.Debug("Failed to checkImportTx", []logger.Field{
								{"txID", utxo.TxID},
								{"error", err}}...)
							return
						}

						if ok {
							continue
						}

						eh.Logger.Debug("No reward UTXO found for txID", []logger.Field{{"txID", utxo.TxID}}...)
						continue
					}

					dur := rand.Intn(10000)
					time.Sleep(time.Millisecond * time.Duration(dur)) // sleep because concurrent reportUTXO can cause failure.

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
			err = errors.Errorf("failed to report UTXO, tx receipt:%v", rcpt)
			return err
		}

		newTxHash := tx.Hash()
		txHash = &newTxHash
		return nil
	})
	return
}

func (eh *UTXOTracker) checkRewardUTXO(ctx context.Context, txID ids.ID) (bool, error) {
	utxoBytesArr, err := eh.PChainClient.GetRewardUTXOs(ctx, &api.GetTxArgs{TxID: txID})
	if err != nil {
		return false, errors.Wrapf(err, "failed to get reward UTXO for txID %q", txID)
	}

	if len(utxoBytesArr) == 0 {
		return false, nil
	}

	var utxos []*avax.UTXO
	for _, utxoBytes := range utxoBytesArr {
		if utxoBytes == nil {
			continue
		}
		utxo := &avax.UTXO{}
		if version, err := platformvm.Codec.Unmarshal(utxoBytes, utxo); err != nil {
			return false, errors.Wrapf(err, "error parsing UTXO, codec version:%v", version)
		}
		utxos = append(utxos, utxo)
	}

	if len(utxos) == 0 {
		return false, nil
	}
	return true, nil
}

func (eh *UTXOTracker) checkImportTx(ctx context.Context, txID ids.ID) (bool, error) {
	txBytes, err := eh.PChainClient.GetTx(ctx, txID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to GetTx")
	}
	tx := &txs.Tx{}
	if version, err := platformvm.Codec.Unmarshal(txBytes, tx); err != nil {
		return false, errors.Wrapf(err, "error parsing tx, codec version:%v", version)
	}
	_, ok := tx.Unsigned.(*txs.ImportTx)
	if ok {
		return true, nil
	}
	return false, nil
}
