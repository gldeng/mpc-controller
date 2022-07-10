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

// Emit event: *events.RewardedStakeReportedEvent

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
				eh.reportRewardedStake(ctx)
			}()
		})

		eh.lock.Lock()
		eh.utxoFetchedEvtObjMap[evt.AddDelegatorTxID.String()] = evtObj
		eh.lock.Unlock()
	}
}

func (eh *UTXOReporter) reportRewardedStake(ctx context.Context) {
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

				txHash, err := eh.retryReportRewardedStake(ctx, groupIDBytes, myIndex, dnmPubKeyBtes, evt.AddDelegatorTxID)
				if err != nil {
					eh.Logger.Error("Failed to report rewarded stake request", []logger.Field{
						{"error", err},
						{"addDelegatorTxID", bytes.Bytes32ToHex(evt.AddDelegatorTxID)},
						{"txHash", txHash}}...)
					break
				}

				newEvt := &events.RewardedStakeReportedEvent{
					AddDelegatorTxID: evt.AddDelegatorTxID,
					PubKeyHex:        genPubKeyInfo.GenPubKeyHex,
					GroupIDHex:       genPubKeyInfo.GroupIdHex,
					MyIndex:          myIndex,
					TxHash:           txHash,
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

func (eh *UTXOReporter) retryReportRewardedStake(ctx context.Context, groupId [32]byte, myIndex *big.Int, publicKey []byte, txID [32]byte) (txHash *common.Hash, err error) {
	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var tx *types.Transaction

	backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
		tx, err = transactor.ReportRewardedStake(eh.Signer, groupId, myIndex, publicKey, txID)
		if err != nil {
			err = errors.Wrapf(err, "failed to report rewarded stake for addDelegatorTxID:%v", bytes.Bytes32ToHex(txID))
			return err
		}

		time.Sleep(time.Second * 3)

		var rcpt *types.Receipt
		rcpt, err = eh.Receipter.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			return errors.WithStack(err)
		}

		if rcpt.Status != 1 {
			err = errors.Errorf("reporting rewarded stake transaction for addDelegatorTxID %q failed", bytes.Bytes32ToHex(txID))
			return err
		}

		newTxHash := tx.Hash()
		txHash = &newTxHash
		return nil
	})
	return
}
