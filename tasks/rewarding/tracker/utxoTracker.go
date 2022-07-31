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
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/avalido/mpc-controller/utils/work"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
	"sync"
	"time"
)

const (
	checkUTXOInterval = time.Second * 1
)

// Subscribe event: *events.ReportedGenPubKeyEvent

// Publish event: *events.UTXOReportedEvent

type UTXOTracker struct {
	ContractAddr common.Address
	Logger       logger.Logger
	PChainClient platformvm.Client
	Publisher    dispatcher.Publisher
	Receipter    chain.Receipter
	Signer       *bind.TransactOpts
	Transactor   bind.ContractTransactor

	reportedGenPubKeyEventChan  chan *events.ReportedGenPubKeyEvent
	reportedGenPubKeyEventCache map[ids.ShortID]*events.ReportedGenPubKeyEvent

	reportUTXOsChan          chan *reportUTXOs
	reportUTXOTaskAddedCache *ristretto.Cache

	UTXOsFetchedEventCache *ristretto.Cache
	utxoReportedEventCache *ristretto.Cache
	UTXOExportedEventCache *ristretto.Cache

	reportUTXOWs *work.Workshop

	once sync.Once
	lock sync.Mutex
}

type reportUTXOs struct {
	utxos      []*avax.UTXO
	groupID    [32]byte
	partiIndex *big.Int
	genPubKey  []byte
}

type reportUTXO struct {
	utxo       *avax.UTXO
	groupID    [32]byte `copier:"must"`
	partiIndex *big.Int `copier:"must"`
	genPubKey  []byte   `copier:"must"`
}

func (eh *UTXOTracker) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.reportedGenPubKeyEventChan = make(chan *events.ReportedGenPubKeyEvent, 1024)
		eh.reportedGenPubKeyEventCache = make(map[ids.ShortID]*events.ReportedGenPubKeyEvent)

		eh.reportUTXOWs = work.NewWorkshop(eh.Logger, "reportUTXOs", time.Minute*10, 10)
		eh.reportUTXOsChan = make(chan *reportUTXOs, 256)

		reportUTXOTaskAddedCache, _ := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		eh.reportUTXOTaskAddedCache = reportUTXOTaskAddedCache

		utxoReportedEventCache, _ := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		eh.utxoReportedEventCache = utxoReportedEventCache

		go eh.fetchUTXOs(ctx)
		go eh.reportUTXOs(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *events.ReportedGenPubKeyEvent:
		eh.reportedGenPubKeyEventCache[evt.PChainAddress] = evt
	}
}

func (eh *UTXOTracker) fetchUTXOs(ctx context.Context) {
	t := time.NewTicker(checkUTXOInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for addr, pubKeyEvt := range eh.reportedGenPubKeyEventCache {
				utxosFromStake, err := eh.requestUTXOsFromStake(ctx, addr)
				if err != nil {
					eh.Logger.DebugOnError(err, "Failed to fetch UTXOs",
						[]logger.Field{{"pChainAddr", addr}}...)
					continue
				}

				groupIdBytes := bytes.HexTo32Bytes(pubKeyEvt.GroupIdHex)
				partiIndex := pubKeyEvt.MyPartiIndex
				genPubKeyBytes := bytes.HexToBytes(pubKeyEvt.GenPubKeyHex)

				var unreportedUTXOsFromStake []*avax.UTXO
				for _, utxo := range utxosFromStake {
					utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
					_, ok := eh.reportUTXOTaskAddedCache.Get(utxoID)
					if ok {
						continue
					}
					_, ok = eh.utxoReportedEventCache.Get(utxoID)
					if ok {
						continue
					}
					_, ok = eh.UTXOExportedEventCache.Get(utxoID)
					if ok {
						continue
					}
					unreportedUTXOsFromStake = append(unreportedUTXOsFromStake, utxo)
				}

				if len(unreportedUTXOsFromStake) == 0 {
					continue
				}

				reportUTXOs := &reportUTXOs{
					utxos:      unreportedUTXOsFromStake,
					groupID:    groupIdBytes,
					partiIndex: partiIndex,
					genPubKey:  genPubKeyBytes,
				}

				select {
				case <-ctx.Done():
					return
				case eh.reportUTXOsChan <- reportUTXOs:
				}
			}
		}
	}
}

func (eh *UTXOTracker) reportUTXOs(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case reportUTXOs := <-eh.reportUTXOsChan:
			for _, utxo := range reportUTXOs.utxos {
				reportUtxo := &reportUTXO{
					utxo:       utxo,
					groupID:    reportUTXOs.groupID,
					partiIndex: reportUTXOs.partiIndex,
					genPubKey:  reportUTXOs.genPubKey,
				}

				reportEvt := &events.UTXOReportedEvent{ // to be shared with utxoPorter timely.
					NativeUTXO:     utxo,
					MpcUTXO:        myAvax.MpcUTXOFromUTXO(utxo),
					TxHash:         nil,
					GenPubKeyBytes: reportUtxo.genPubKey,
					GroupIDBytes:   reportUtxo.groupID,
					PartiIndex:     reportUtxo.partiIndex,
				}
				utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
				eh.UTXOsFetchedEventCache.SetWithTTL(utxoID, reportEvt, 1, time.Hour)
				eh.UTXOsFetchedEventCache.Wait()

				eh.reportUTXOTaskAddedCache.SetWithTTL(utxoID, " ", 1, time.Hour)
				eh.reportUTXOTaskAddedCache.Wait()

				eh.reportUTXOWs.AddTask(ctx, &work.Task{
					Args: reportUtxo,
					Ctx:  ctx,
					WorkFns: []work.WorkFn{func(ctx context.Context, args interface{}) {
						reportUtxo := args.(*reportUTXO)
						utxo := reportUtxo.utxo

						utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
						_, ok := eh.utxoReportedEventCache.Get(utxoID)
						if ok {
							return
						}
						_, ok = eh.UTXOExportedEventCache.Get(utxoID)
						if ok {
							return
						}
						txHash, err := eh.doReportUTXO(ctx, reportUtxo)
						if err != nil {
							eh.Logger.ErrorOnError(err, "Failed to reportEvt UTXO", []logger.Field{{"utxoID", utxo}}...)
							return
						}

						reportEvt := &events.UTXOReportedEvent{
							NativeUTXO:     utxo,
							MpcUTXO:        myAvax.MpcUTXOFromUTXO(utxo),
							TxHash:         txHash,
							GenPubKeyBytes: reportUtxo.genPubKey,
							GroupIDBytes:   reportUtxo.groupID,
							PartiIndex:     reportUtxo.partiIndex,
						}

						eh.utxoReportedEventCache.SetWithTTL(utxoID, reportEvt, 1, time.Hour)
						eh.utxoReportedEventCache.Wait()

						//eh.Publisher.Publish(ctx, dispatcher.NewEvtObj(reportEvt, nil))
						switch utxo.OutputIndex {
						case 0:
							eh.Logger.Debug("Principal UTXO REPORTED", []logger.Field{{"utxoID", utxo.UTXOID}}...)
						case 1:
							eh.Logger.Debug("Reward UTXO REPORTED", []logger.Field{{"utxoID", utxo.UTXOID}}...)
						}
					}},
				})
			}
		}
	}
}

func (eh *UTXOTracker) requestUTXOsFromStake(ctx context.Context, addr ids.ShortID) (utxos []*avax.UTXO, err error) {
	rawUtxos, err := eh.getUTXOs(ctx, addr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get native UTXOs")
	}

	if len(rawUtxos) == 0 {
		return nil, nil
	}

	var filterUtxos []*avax.UTXO

	for _, utxo := range rawUtxos {
		ok, err := eh.isFromImportTx(ctx, utxo.TxID)
		if err != nil {
			eh.Logger.Debug("Failed to check whether UTXO is from importTX", []logger.Field{
				{"txID", utxo.TxID},
				{"error", err}}...)
			continue
		}

		// An address dedicated to delegate stake may receive three kinds of native UTXO on P-Chain,
		// namely atomic importTx UTXO, principal UTXO and reward UTXO after stake period end.
		// ImportTx UTXOs serves for addDelegatorTx in our program and should be excluded from exporting its avax to C-Chain
		// todo: confirm whether there other potential UTXO that should be excluded, too.
		if ok {
			continue
		}

		filterUtxos = append(filterUtxos, utxo)
	}
	return filterUtxos, nil
}

func (eh *UTXOTracker) getUTXOs(ctx context.Context, addr ids.ShortID) (utxos []*avax.UTXO, err error) {
	var utxoBytesArr [][]byte
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		utxoBytesArr, _, _, err = eh.PChainClient.GetUTXOs(ctx, []ids.ShortID{addr}, 0, addr, ids.ID{})
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to get UTXOs")
	if err != nil {
		return
	}

	utxos, err = myAvax.ParseUTXOs(utxoBytesArr)
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

func (eh *UTXOTracker) doReportUTXO(ctx context.Context, reportUtxo *reportUTXO) (txHash *common.Hash, err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor) // todo: extract to reuse in multi flows.
	if err != nil {
		return nil, errors.WithStack(err)
	}

	utxo := reportUtxo.utxo

	var tx *types.Transaction
	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
		tx, err = transactor.ReportUTXO(eh.Signer, reportUtxo.groupID, reportUtxo.partiIndex, reportUtxo.genPubKey, utxo.TxID, utxo.OutputIndex)
		if err != nil {
			return true, errors.WithStack(err)
		}
		var rcpt *types.Receipt
		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
			rcpt, err = eh.Receipter.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				return true, errors.WithStack(err) // todo: consider reverted case
			}
			if rcpt.Status != 1 {
				return true, errors.Errorf("tx receipt status != 1")
			}
			newTxHash := tx.Hash()
			txHash = &newTxHash
			return false, nil
		})

		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, nil
	})
	err = errors.Wrapf(err, "failed to report UTXO, txID:%v, outputIndex:%v", bytes.Bytes32ToHex(utxo.TxID), utxo.OutputIndex)
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

func (eh *UTXOTracker) isFromImportTx(ctx context.Context, txID ids.ID) (bool, error) {
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
