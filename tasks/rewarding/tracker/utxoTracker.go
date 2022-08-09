package tracker

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/avalido/mpc-controller/utils/work"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
	"sync"
	"time"
)

const (
	checkUTXOInterval = time.Second * 1
)

// Subscribe event: *events.KeyGenerated

// Publish event: *events.UTXOReported

type UTXOTracker struct {
	ContractAddr common.Address
	Logger       logger.Logger
	PChainClient platformvm.Client
	Publisher    dispatcher.Publisher
	PubKey       []byte
	DB           storage.DB
	Transactor   transactor.Transactor
	//Receipter    chain.Receipter
	//Signer       *bind.TransactOpts
	//Transactor   bind.ContractTransactor

	//genPubKeyChan  chan *events.KeyGenerated
	genPubKeyCache map[ids.ShortID]*events.KeyGenerated

	utxosToExportCh              chan *utxosToExport
	joinExportUTXOTaskAddedCache *ristretto.Cache

	UTXOsFetchedEventCache *ristretto.Cache
	utxoReportedEventCache *ristretto.Cache
	UTXOExportedEventCache *ristretto.Cache

	joinUTXOExportWs *work.Workshop

	once sync.Once
	lock sync.Mutex
}

type utxosToExport struct {
	utxos []*avax.UTXO
	//groupID    [32]byte
	//partiIndex *big.Int
	genPubKey storage.PubKey
}

type utxoToExport struct {
	utxo       *avax.UTXO
	groupID    [32]byte
	partiIndex *big.Int
	genPubKey  []byte
}

func (eh *UTXOTracker) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		//eh.genPubKeyChan = make(chan *events.KeyGenerated)
		eh.genPubKeyCache = make(map[ids.ShortID]*events.KeyGenerated)

		eh.joinUTXOExportWs = work.NewWorkshop(eh.Logger, "utxosToExport", time.Minute*10, 10)
		eh.utxosToExportCh = make(chan *utxosToExport, 256)

		reportUTXOTaskAddedCache, _ := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		eh.joinExportUTXOTaskAddedCache = reportUTXOTaskAddedCache

		utxoReportedEventCache, _ := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		eh.utxoReportedEventCache = utxoReportedEventCache

		go eh.fetchUTXOs(ctx)
		go eh.joinExportUTXOs(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *events.KeyGenerated:
		genPubKey := storage.PubKey(evt.PublicKey)
		pChainAddr, err := genPubKey.PChainAddress()
		if err != nil {
			eh.Logger.ErrorOnError(err, "Failed to get P-Chain address")
			break
		}
		eh.genPubKeyCache[pChainAddr] = evt
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
			for pChainAddr, genPubKey := range eh.genPubKeyCache {
				utxosFromStake, err := eh.requestUTXOsFromStake(ctx, pChainAddr)
				if err != nil {
					eh.Logger.DebugOnError(err, "Failed to fetch UTXOs",
						[]logger.Field{{"pChainAddr", pChainAddr}}...)
					continue
				}

				var nativeUTXOsToExport []*avax.UTXO
				for _, utxo := range utxosFromStake {
					utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
					_, ok := eh.joinExportUTXOTaskAddedCache.Get(utxoID)
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
					nativeUTXOsToExport = append(nativeUTXOsToExport, utxo)
				}

				if len(nativeUTXOsToExport) == 0 {
					continue
				}

				utxosToExport := &utxosToExport{
					utxos: nativeUTXOsToExport,
					//groupID:    groupIdBytes,
					//partiIndex: partiIndex,
					genPubKey: genPubKey.PublicKey,
				}

				select {
				case <-ctx.Done():
					return
				case eh.utxosToExportCh <- utxosToExport:
				}
			}
		}
	}
}

func (eh *UTXOTracker) joinExportUTXOs(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case utxosToExport := <-eh.utxosToExportCh:
			for _, utxo := range utxosToExport.utxos {
				genPubKey := storage.GeneratedPublicKey{
					GenPubKey: utxosToExport.genPubKey,
				}
				err := eh.DB.LoadModel(ctx, &genPubKey)
				if err != nil {
					eh.Logger.ErrorOnError(err, "failed to load generated public key", []logger.Field{{"key", genPubKey.Key()}}...)
					break
				}

				participant := storage.Participant{
					PubKey:  hash256.FromBytes(eh.PubKey),
					GroupId: genPubKey.GroupId,
				}

				err = eh.DB.LoadModel(ctx, &participant)
				if err != nil {
					eh.Logger.ErrorOnError(err, "failed to load participant", []logger.Field{{"key", participant.Key()}}...)
					break
				}

				exportUTXOReq := storage.ExportUTXORequest{
					TxID:        utxo.TxID,
					OutputIndex: utxo.OutputIndex,
					GenPubKey:   genPubKey.GenPubKey,
				}

				partiId := participant.ParticipantId()
				reqHash := exportUTXOReq.ReqHash()

				joinReq := storage.JoinRequest{
					ReqHash: reqHash,
					PartiId: partiId,
					Args:    &exportUTXOReq,
				}

				//reportUtxo := &utxoToExport{
				//	utxo:       utxo,
				//	//groupID:    utxosToExport.groupID,
				//	//partiIndex: utxosToExport.partiIndex,
				//	genPubKey: utxosToExport.genPubKey,
				//}
				//
				//reportEvt := &events.UTXOReported{ // to be shared with utxoPorter timely.
				//	NativeUTXO:     utxo,
				//	MpcUTXO:        myAvax.MpcUTXOFromUTXO(utxo),
				//	TxHash:         nil,
				//	GenPubKeyBytes: reportUtxo.genPubKey,
				//	GroupIDBytes:   reportUtxo.groupID,
				//	PartiIndex:     reportUtxo.partiIndex,
				//}
				//utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
				//eh.UTXOsFetchedEventCache.SetWithTTL(utxoID, reportEvt, 1, time.Hour)
				//eh.UTXOsFetchedEventCache.Wait()
				//
				//eh.joinExportUTXOTaskAddedCache.SetWithTTL(utxoID, " ", 1, time.Hour)
				//eh.joinExportUTXOTaskAddedCache.Wait()

				eh.joinUTXOExportWs.AddTask(ctx, &work.Task{
					Args: &joinReq,
					Ctx:  ctx,
					WorkFns: []work.WorkFn{
						func(ctx context.Context, args interface{}) {
							joinReq := args.(*storage.JoinRequest)
							partiId := joinReq.PartiId
							reqHash := joinReq.ReqHash
							_, _, err := eh.Transactor.JoinRequest(ctx, partiId, reqHash)
							if err != nil {
								switch errors.Cause(err).(type) {
								case *transactor.ErrTypQuorumAlreadyReached:
									eh.Logger.DebugOnError(err, "Join UTXO export request not accepted", []logger.Field{{"reqHash", reqHash}}...)
								case *transactor.ErrTypAttemptToRejoin:
									eh.Logger.DebugOnError(err, "Join UTXO export request not accepted", []logger.Field{{"reqHash", reqHash}}...)
								default:
									eh.Logger.ErrorOnError(err, "Failed to join UTXO export request", []logger.Field{{"reqHash", reqHash}}...)
								}
								return
							}

							err = eh.DB.SaveModel(ctx, joinReq)
							eh.Logger.ErrorOnError(err, "Failed to save JoinRequest for UTXO export", []logger.Field{{"joinReq", joinReq}}...)
							return
							//reportUtxo := args.(*utxoToExport)
							//utxo := reportUtxo.utxo
							//
							//utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
							//_, ok := eh.utxoReportedEventCache.Get(utxoID)
							//if ok {
							//	return
							//}
							//_, ok = eh.UTXOExportedEventCache.Get(utxoID)
							//if ok {
							//	return
							//}
							//txHash, err := eh.doReportUTXO(ctx, reportUtxo)
							//if err != nil {
							//	eh.Logger.ErrorOnError(err, "Failed to reportEvt UTXO", []logger.Field{{"utxoID", utxo}}...)
							//	return
							//}
							//
							//reportEvt := &events.UTXOReported{
							//	NativeUTXO:     utxo,
							//	MpcUTXO:        myAvax.MpcUTXOFromUTXO(utxo),
							//	TxHash:         txHash,
							//	GenPubKeyBytes: reportUtxo.genPubKey,
							//	GroupIDBytes:   reportUtxo.groupID,
							//	PartiIndex:     reportUtxo.partiIndex,
							//}
							//
							//eh.utxoReportedEventCache.SetWithTTL(utxoID, reportEvt, 1, time.Hour)
							//eh.utxoReportedEventCache.Wait()
							//
							////eh.Publisher.Publish(ctx, dispatcher.NewEvtObj(reportEvt, nil))
							//switch utxo.OutputIndex {
							//case 0:
							//	eh.Logger.Debug("Principal UTXO REPORTED", []logger.Field{{"utxoID", utxo.UTXOID}}...)
							//case 1:
							//	eh.Logger.Debug("Reward UTXO REPORTED", []logger.Field{{"utxoID", utxo.UTXOID}}...)
						},
					},
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

//func (eh *UTXOTracker) doReportUTXO(ctx context.Context, reportUtxo *utxoToExport) (txHash *common.Hash, err error) {
//	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor) // todo: extract to reuse in multi flows.
//	if err != nil {
//		return nil, errors.WithStack(err)
//	}
//
//	utxo := reportUtxo.utxo
//
//	var tx *types.Transaction
//	err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
//		tx, err = transactor.ReportUTXO(eh.Signer, reportUtxo.groupID, reportUtxo.partiIndex, reportUtxo.genPubKey, utxo.TxID, utxo.OutputIndex)
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//		var rcpt *types.Receipt
//		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (bool, error) {
//			rcpt, err = eh.Receipter.TransactionReceipt(ctx, tx.Hash())
//			if err != nil {
//				return true, errors.WithStack(err) // todo: consider reverted case
//			}
//			if rcpt.Status != 1 {
//				return true, errors.Errorf("tx receipt status != 1")
//			}
//			newTxHash := tx.Hash()
//			txHash = &newTxHash
//			return false, nil
//		})
//
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//		return false, nil
//	})
//	err = errors.Wrapf(err, "failed to report UTXO, txID:%v, outputIndex:%v", bytes.Bytes32ToHex(utxo.TxID), utxo.OutputIndex)
//	return
//}

//func (eh *UTXOTracker) checkRewardUTXO(ctx context.Context, txID ids.ID) (bool, error) {
//	utxoBytesArr, err := eh.PChainClient.GetRewardUTXOs(ctx, &api.GetTxArgs{TxID: txID})
//	if err != nil {
//		return false, errors.Wrapf(err, "failed to get reward UTXO for txID %q", txID)
//	}
//
//	if len(utxoBytesArr) == 0 {
//		return false, nil
//	}
//
//	var utxos []*avax.UTXO
//	for _, utxoBytes := range utxoBytesArr {
//		if utxoBytes == nil {
//			continue
//		}
//		utxo := &avax.UTXO{}
//		if version, err := platformvm.Codec.Unmarshal(utxoBytes, utxo); err != nil {
//			return false, errors.Wrapf(err, "error parsing UTXO, codec version:%v", version)
//		}
//		utxos = append(utxos, utxo)
//	}
//
//	if len(utxos) == 0 {
//		return false, nil
//	}
//	return true, nil
//}

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
