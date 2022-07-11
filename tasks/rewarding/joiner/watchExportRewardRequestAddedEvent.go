package joiner

//import (
//	"context"
//	"github.com/ava-labs/avalanchego/ids"
//	"github.com/avalido/mpc-controller/cache"
//	"github.com/avalido/mpc-controller/chain"
//	"github.com/avalido/mpc-controller/contract"
//	"github.com/avalido/mpc-controller/dispatcher"
//	"github.com/avalido/mpc-controller/events"
//	"github.com/avalido/mpc-controller/logger"
//	"github.com/avalido/mpc-controller/utils/backoff"
//	"github.com/ethereum/go-ethereum/accounts/abi/bind"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/event"
//	"github.com/pkg/errors"
//	"math/big"
//	"strings"
//	"sync"
//	"time"
//)
//
//// Accept event: *events.ContractFiltererCreatedEvent
//// Accept event: *events.ReportedGenPubKeyEvent
//
//// Emit event: *contract.ExportUTXORequestAddedEvent
//// Emit event: *events.JoinedExportUTXORequestEvent
//
//type ExportUTXORequestJoiner struct {
//	Cache           cache.ICache
//	ContractAddr    common.Address
//	Logger          logger.Logger
//	MyPubKeyHashHex string
//	Publisher       dispatcher.Publisher
//	Receipter       chain.Receipter
//	Signer          *bind.TransactOpts
//	Transactor      bind.ContractTransactor
//
//	once               sync.Once
//	lock               sync.Mutex
//	genPubKeyEvtObjMap map[ids.ShortID]*dispatcher.EventObject // todo: persistence and restore
//
//	pubKeyBytes [][]byte
//	filterer    bind.ContractFilterer
//	sub         event.Subscription
//	sink        chan *contract.MpcManagerExportUTXORequestAdded
//	done        chan struct{}
//}
//
//func (eh *ExportUTXORequestJoiner) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
//	switch evt := evtObj.Event.(type) {
//	case *events.ContractFiltererCreatedEvent:
//		eh.filterer = evt.Filterer
//	case *events.ReportedGenPubKeyEvent:
//		eh.once.Do(func() {
//			eh.genPubKeyEvtObjMap = make(map[ids.ShortID]*dispatcher.EventObject)
//			go func() {
//				eh.watchAndJoinExportUTXORequest(ctx)
//			}()
//		})
//
//		eh.lock.Lock()
//		eh.genPubKeyEvtObjMap[evt.PChainAddress] = evtObj
//		eh.lock.Unlock()
//
//		//dnmPubKeyBtes, err := crypto.DenormalizePubKeyFromHex(evt.Val.CompressedGenPubKeyHex)
//		//if err != nil {
//		//	eh.Logger.Error("Failed to denormalized generated public key", []logger.Field{{"error", err}}...)
//		//	break
//		//}
//		//
//		//eh.pubKeyBytes = append(eh.pubKeyBytes, dnmPubKeyBtes)
//	}
//	//if len(eh.pubKeyBytes) > 0 {
//	//	eh.doWatchExportRewardRequestAdded(evtObj.Context)
//	//}
//}
//
//func (eh *ExportUTXORequestJoiner) watchAndJoinExportUTXORequest(ctx context.Context) {
//
//}
//
//func (eh *ExportUTXORequestJoiner) doWatchExportRewardRequestAdded(ctx context.Context) {
//	newSink := make(chan *contract.MpcManagerExportUTXORequestAdded)
//	err := eh.subscribeExportRewardRequestAdded(ctx, newSink, eh.pubKeyBytes)
//	if err == nil {
//		eh.sink = newSink
//		if eh.done != nil {
//			close(eh.done)
//		}
//		eh.done = make(chan struct{})
//		eh.watchAndJoinExportUTXORequest(ctx)
//	}
//}
//
//func (eh *ExportUTXORequestJoiner) subscribeExportRewardRequestAdded(ctx context.Context, sink chan<- *contract.MpcManagerExportUTXORequestAdded, pubKeys [][]byte) error {
//	if eh.sub != nil {
//		eh.sub.Unsubscribe()
//	}
//
//	err := backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
//		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
//		if err != nil {
//			eh.Logger.Error("Failed to create MpcManagerFilterer", []logger.Field{{"error", err}}...)
//			return errors.WithStack(err)
//		}
//
//		newSub, err := filter.WatchExportUTXORequestAdded(nil, sink, pubKeys)
//		if err != nil {
//			eh.Logger.Error("Failed to watch ExportRewardRequestStarted event", []logger.Field{{"error", err}}...)
//			return errors.WithStack(err)
//		}
//
//		eh.sub = newSub
//		return nil
//	})
//
//	return err
//}
//
//func (eh *ExportUTXORequestJoiner) joinExportUTXORequest(ctx context.Context) {
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-eh.done:
//				return
//			case evt := <-eh.sink:
//				transformedEvt := events.ExportUTXORequestAddedEvent{
//					TxID: evt.RewaredStakeTxId,
//					PublicKeyHash:    evt.PublicKey,
//					TxHash:           evt.Raw.TxHash,
//				}
//				evtObj := dispatcher.NewRootEventObject("ExportUTXORequestJoiner", transformedEvt, ctx)
//				eh.Publisher.Publish(ctx, evtObj)
//			case err := <-eh.sub.Err():
//				eh.Logger.ErrorOnError(err, "Got an error during watching ExportRewardRequestAdded event", []logger.Field{{"error", err}}...)
//			}
//		}
//	}()
//}
//
//func (eh *ExportUTXORequestJoiner) doJoinExportUTXORequest(ctx context.Context, groupId [32]byte, partiIndex *big.Int, genPubKey []byte, utxoTxID [32]byte, utxoOutputIndex uint32) (txHash *common.Hash, err error) {
//	transactor, err := contract.NewMpcManagerTransactor(eh.ContractAddr, eh.Transactor)
//	if err != nil {
//		return nil, errors.WithStack(err)
//	}
//
//	var tx *types.Transaction
//
//	err = backoff.RetryFnExponentialForever(eh.Logger, ctx, func() error {
//		tx, err = transactor.JoinExportUTXO(eh.Signer, groupId, partiIndex, genPubKey, utxoTxID, utxoOutputIndex)
//		if err != nil {
//			if strings.Contains(err.Error(), "execution reverted: Cannot join anymore") {
//				tx = nil
//				eh.Logger.Debug("Cannot join anymore")
//				return nil
//			}
//			err = errors.Wrap(err, "failed to join request.")
//			return err
//		}
//
//		time.Sleep(time.Second * 3)
//
//		var rcpt *types.Receipt
//		rcpt, err = eh.Receipter.TransactionReceipt(ctx, tx.Hash())
//		if err != nil {
//			return errors.WithStack(err)
//		}
//
//		if rcpt.Status != 1 {
//			err = errors.New("Transaction failed")
//			return err
//		}
//
//		newTxHash := tx.Hash()
//		txHash = &newTxHash
//		return nil
//	})
//	return
//}
