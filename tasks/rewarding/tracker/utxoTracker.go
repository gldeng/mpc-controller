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
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"time"
)

const (
	checkUTXOInterval = time.Second * 1
)

// Subscribe event: *events.KeyGenerated

type UTXOTracker struct {
	BoundTransactor        transactor.Transactor
	ClientPChain           platformvm.Client
	DB                     storage.DB
	Logger                 logger.Logger
	PartiPubKey            storage.PubKey
	UTXOExportedEventCache *ristretto.Cache
	UTXOsFetchedEventCache *ristretto.Cache

	genPubKeyCache               map[ids.ShortID]*events.KeyGenerated
	utxosToExportCh              chan *utxosToExport
	joinUTXOExportTaskAddedCache *ristretto.Cache
	joinedUTXOExportCache        *ristretto.Cache
	joinUTXOExportWs             *work.Workshop
	once                         sync.Once
}

type utxosToExport struct {
	utxos     []*avax.UTXO
	genPubKey storage.PubKey
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
		eh.joinUTXOExportTaskAddedCache = reportUTXOTaskAddedCache

		utxoReportedEventCache, _ := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		eh.joinedUTXOExportCache = utxoReportedEventCache

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
					_, ok := eh.joinUTXOExportTaskAddedCache.Get(utxoID)
					if ok {
						continue
					}
					_, ok = eh.joinedUTXOExportCache.Get(utxoID)
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
					utxos:     nativeUTXOsToExport,
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
					PubKey:  hash256.FromBytes(eh.PartiPubKey),
					GroupId: genPubKey.GroupId,
				}

				err = eh.DB.LoadModel(ctx, &participant)
				if err != nil {
					eh.Logger.ErrorOnError(err, "failed to load participant", []logger.Field{{"key", participant.Key()}}...)
					break
				}

				exportUTXOReq := storage.ExportUTXORequest{
					TxID:               utxo.TxID,
					OutputIndex:        utxo.OutputIndex,
					GeneratedPublicKey: &genPubKey,
				}

				partiId := participant.ParticipantId()
				reqHash := exportUTXOReq.ReqHash()

				joinReq := storage.JoinRequest{
					ReqHash: reqHash,
					PartiId: partiId,
					Args:    &exportUTXOReq,
				}

				utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
				eh.UTXOsFetchedEventCache.SetWithTTL(utxoID, utxo, 1, time.Hour)
				eh.UTXOsFetchedEventCache.Wait()
				eh.joinUTXOExportTaskAddedCache.SetWithTTL(utxoID, " ", 1, time.Hour)
				eh.joinUTXOExportTaskAddedCache.Wait()

				eh.joinUTXOExportWs.AddTask(ctx, &work.Task{
					Args: &joinReq,
					Ctx:  ctx,
					WorkFns: []work.WorkFn{
						func(ctx context.Context, args interface{}) {
							joinReq := args.(*storage.JoinRequest)
							partiId := joinReq.PartiId
							reqHash := joinReq.ReqHash
							reqArgs := joinReq.Args.(*storage.ExportUTXORequest)
							txID := reqArgs.TxID
							outputIndex := reqArgs.OutputIndex

							exportUTXOReq := joinReq.Args.(*storage.ExportUTXORequest)
							utxoID := exportUTXOReq.TxID.String() + strconv.Itoa(int(exportUTXOReq.OutputIndex))
							_, ok := eh.joinedUTXOExportCache.Get(utxoID)
							if ok {
								return
							}
							_, ok = eh.UTXOExportedEventCache.Get(utxoID)
							if ok {
								return
							}

							if err = eh.DB.SaveModel(ctx, joinReq); err != nil {
								eh.Logger.ErrorOnError(err, "Failed to save JoinRequest for UTXO export", []logger.Field{{"joinReq", joinReq}}...)
								return
							}

							_, _, err := eh.BoundTransactor.JoinRequest(ctx, partiId, reqHash)
							if err != nil {
								switch errors.Cause(err).(type) {
								case *transactor.ErrTypQuorumAlreadyReached:
									eh.Logger.DebugOnError(err, "Join UTXO export request not accepted", []logger.Field{{"reqHash", reqHash.String()}, {"txID", txID}, {"outputIndex", outputIndex}}...)
								case *transactor.ErrTypAttemptToRejoin:
									eh.Logger.DebugOnError(err, "Join UTXO export request not accepted", []logger.Field{{"reqHash", reqHash.String()}, {"txID", txID}, {"outputIndex", outputIndex}}...)
								case *transactor.ErrTypExecutionReverted:
									eh.Logger.DebugOnError(err, "Join UTXO export request not accepted", []logger.Field{{"reqHash", reqHash.String()}, {"txID", txID}, {"outputIndex", outputIndex}}...)
								default:
									eh.Logger.ErrorOnError(err, "Failed to join UTXO export request", []logger.Field{{"reqHash", reqHash.String()}, {"txID", txID}, {"outputIndex", outputIndex}}...)
								}
								return
							}

							eh.joinedUTXOExportCache.SetWithTTL(utxoID, " ", 1, time.Hour)
							eh.joinedUTXOExportCache.Wait()

							switch outputIndex {
							case 0:
								eh.Logger.Debug("Joined principal UTXO export request", []logger.Field{{"reqHash", reqHash.String()}, {"txID", txID}, {"outputIndex", outputIndex}}...)
							case 1:
								eh.Logger.Debug("Joined reward UTXO export request", []logger.Field{{"reqHash", reqHash.String()}, {"txID", txID}, {"outputIndex", outputIndex}}...)
							}
							return
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
		utxoBytesArr, _, _, err = eh.ClientPChain.GetUTXOs(ctx, []ids.ShortID{addr}, 0, addr, ids.ID{})
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

func (eh *UTXOTracker) isFromImportTx(ctx context.Context, txID ids.ID) (bool, error) {
	txBytes, err := eh.ClientPChain.GetTx(ctx, txID)
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
