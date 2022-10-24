package utxotracker

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	myAvax "github.com/avalido/mpc-controller/utils/port/avax"
	"github.com/dgraph-io/ristretto"
	kbcevents "github.com/kubecost/events"
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"time"
)

const (
	checkUTXOInterval = time.Second * 15
)

// Subscribe event: *events.KeyGenerated

type UTXOTracker struct {
	ClientPChain platformvm.Client
	Logger       logger.Logger

	Cache *ristretto.Cache

	Dispatcher kbcevents.Dispatcher[*events.UTXOFetched]

	genPubKeyCache map[ids.ShortID]*events.KeyGenerated
	once           sync.Once
}

// todo: apply kubecost/events for dispatcher

func (eh *UTXOTracker) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.genPubKeyCache = make(map[ids.ShortID]*events.KeyGenerated)
		go eh.fetchUTXOs(ctx)
	})

	switch evt := evtObj.Event.(type) {
	case *events.KeyGenerated:
		genPubKey := storage.PubKey(evt.PublicKey)
		pChainAddr, err := genPubKey.PChainAddress()
		if err != nil {
			eh.Logger.ErrorOnError(err, "UTXOTracker failed to get P-Chain address")
			break
		}
		eh.genPubKeyCache[pChainAddr] = evt
	}
}

// todo: apply Tx memo for filter

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
					eh.Logger.DebugOnError(err, "UTXOTracker failed to fetch UTXOs")
					continue
				}

				for _, utxo := range utxosFromStake {
					utxoID := utxo.TxID.String() + strconv.Itoa(int(utxo.OutputIndex))
					_, ok := eh.Cache.Get(utxoID)
					if ok {
						continue
					}

					utxoToRecover := events.UTXOFetched{
						UTXO:      utxo,
						GenPubKey: genPubKey.PublicKey,
					}

					eh.Dispatcher.Dispatch(&utxoToRecover)
					eh.Cache.SetWithTTL(utxoID, utxo, 1, time.Hour)
					eh.Cache.Wait()
					eh.Logger.Info("UTXOTracker dispatched UTXOFetched", []logger.Field{{"UTXOFetched", myAvax.MpcUTXOFromUTXO(utxo)}}...)
				}
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
			return nil, errors.Wrapf(err, "failed to check whether UTXO is from importTx")
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
	err = backoff.RetryFnExponential10Times(eh.Logger, ctx, time.Second, time.Second*10, func() (bool, error) {
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
