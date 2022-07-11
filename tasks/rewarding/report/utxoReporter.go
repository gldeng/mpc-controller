package report

import (
	"context"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"math/big"
	"sync"
	"time"
)

// Accept event: *events.UTXOsFetchedEvent

// Emit event: *events.UTXOReportedEvent

type UTXOReporter struct {
	Cache           cache.ICache
	ContractAddr    common.Address
	Logger          logger.Logger
	MyPubKeyHashHex string
	Publisher       dispatcher.Publisher
	Receipter       chain.Receipter
	Signer          *bind.TransactOpts
	Transactor      bind.ContractTransactor

	once                 sync.Once
	lock                 sync.Mutex
	utxoFetchedEvtObjMap map[string]*dispatcher.EventObject // todo: persistence and restore
}

func (eh *UTXOReporter) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.UTXOsFetchedEvent:
		eh.once.Do(func() {
			eh.utxoFetchedEvtObjMap = make(map[string]*dispatcher.EventObject)
			go func() {
				eh.reportUTXO(ctx)
			}()
		})

		eh.lock.Lock()
		eh.utxoFetchedEvtObjMap[evt.AddDelegatorTxID.String()] = evtObj
		eh.lock.Unlock()
	}
}

func (eh *UTXOReporter) reportUTXO(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			eh.lock.Lock()
			for txID, evtObj := range eh.utxoFetchedEvtObjMap {
				evt := evtObj.Event.(*events.UTXOsFetchedEvent)
				dnmGenPubKeyBytes, err := crypto.DenormalizePubKeyFromHex(evt.PubKeyHex) // for Ethereum compatibility
				if err != nil {
					eh.Logger.Error("Failed to DenormalizePubKeyFromHex", []logger.Field{{"err", err}}...)
					break
				}

				dnmGenPubKeyHash := hash256.FromHex(bytes.BytesToHex(dnmGenPubKeyBytes))
				genPubKeyInfo := eh.Cache.GetGeneratedPubKeyInfo(dnmGenPubKeyHash.Hex())
				if genPubKeyInfo == nil {
					eh.Logger.Error("No GeneratedPubKeyInfo found")
					break
				}

				myIndex := eh.Cache.GetMyIndex(eh.MyPubKeyHashHex, dnmGenPubKeyHash.Hex())
				if myIndex == nil {
					eh.Logger.Error("Not found my index.")
					break
				}

				groupIDBytes := bytes.HexTo32Bytes(genPubKeyInfo.GroupIdHex)
				dnmPubKeyBtes, err := crypto.DenormalizePubKeyFromHex(genPubKeyInfo.GenPubKeyHex)
				if err != nil {
					eh.Logger.Error("Failed to denormalized generated public key", []logger.Field{{"error", err}}...)
					break
				}

				txHash, err := eh.reportUTXO(ctx, groupIDBytes, myIndex, dnmPubKeyBtes, evt.AddDelegatorTxID)
				if err != nil {
					eh.Logger.Error("Failed to report rewarded stake request", []logger.Field{
						{"error", err},
						{"addDelegatorTxID", bytes.Bytes32ToHex(evt.AddDelegatorTxID)},
						{"txHash", txHash}}...)
					break
				}

				newEvt := &events.UTXOReportedEvent{
					TxID:           evt.AddDelegatorTxID,
					genPubKeyBytes: genPubKeyInfo.GenPubKeyHex,
					GroupIDBytes:   genPubKeyInfo.GroupIdHex,
					PartiIndex:     myIndex,
					TxHash:         txHash,
				}
				eh.Publisher.Publish(ctx, dispatcher.NewEventObjectFromParent(evtObj, "UTXOReporter", newEvt, evtObj.Context))
				delete(eh.utxoFetchedEvtObjMap, txID)
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func (eh *UTXOReporter) reportUTXO(ctx context.Context, groupId [32]byte, myIndex *big.Int, genPubKey []byte, txID [32]byte, outputIndex uint32) (txHash *common.Hash, err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor) // todo: extract to reuse in multi flows.
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var tx *types.Transaction

	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		tx, err = transactor.ReportUTXO(eh.Signer, groupId, myIndex, genPubKey, txID, outputIndex)
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
