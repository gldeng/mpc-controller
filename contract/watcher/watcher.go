package watcher

import (
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/contract/watcher"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/network/redialer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

// Subscribe event: *events.ContractFiltererCreated
// Subscribe event: *events.GeneratedPubKeyInfoStored

// Process event: *contract.MpcManagerStakeRequestStarted

type Watcher struct {
	Logger       logger.Logger
	PubKeys      [][]byte
	EthWsURL     string
	ContractAddr common.Address
	Publisher    dispatcher.Publisher

	filterer            bind.ContractFilterer
	filtererRecreatedCh chan redialer.Client
	contractFilter      *contract.MpcManagerFilterer
	once                sync.Once

	groupIDs [][32]byte

	participantAddedWatcher   *watcher.Watcher
	keygenRequestAddedWatcher *watcher.Watcher
}

//func (w *Watcher) Init(ctx context.Context) {
//	w.once.Do(func() {
//		reDialer := adapter.EthClientReDialer{
//			L:        logger.Default(),
//			EthURL:        w.EthWsURL,
//			BackOffPolicy: backoff.ExponentialPolicy(0, time.Second, time.Second*10),
//		}
//		ethClient, ethclientCh, err := reDialer.GetClient(context.Background())
//		w.L.FatalOnError(err, "Failed to get eth client")
//		w.filterer = ethClient.(*ethclient.Client)
//		w.filtererRecreatedCh = ethclientCh
//
//		contractFilter, err := contract.NewMpcManagerFilterer(w.ContractAddr, w.filterer)
//		w.L.FatalOnError(err, "Failed to create MpcManagerFilterer")
//		w.contractFilter = contractFilter
//
//		w.watchParticipantAdded(ctx)
//	})
//}
//
//func (w *Watcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
//
//	switch evt := evtObj.Event.(type) {
//	case *events.ParticipantAdded:
//		// watch KeygenRequestAdded
//	case *events.GeneratedPubKeyInfoStored:
//		dnmPubKeyBtes, err := crypto.DenormalizePubKeyFromHex(evt.Val.CompressedGenPubKeyHex)
//		if err != nil {
//			w.L.Error("Failed to denormalized generated public key", []logger.Field{{"error", err}}...)
//			break
//		}
//
//		w.pubKeyBytes = append(w.pubKeyBytes, dnmPubKeyBtes)
//	}
//	if len(w.pubKeyBytes) > 0 {
//		w.doWatchStakeRequestStarted(ctx)
//	}
//}
//
//func (w *Watcher) watchParticipantAdded(ctx context.Context) {
//	mySubPub := ParticipantAddedSubPub(w.GroupIDs)
//	myWatcher := &watcher.Watcher{
//		L:    w.L,
//		Subscribe: mySubPub.Subscribe,
//		Process:   mySubPub.Publish,
//		Filterer:  w.contractFilter,
//		P: w.P,
//	}
//	w.L.FatalOnError(myWatcher.Watch(ctx), "Failed to watch ParticipantAdded")
//	w.participantAddedWatcher = myWatcher
//}
//
//func (w *Watcher) WatchKeygenRequestAdded(ctx context.Context) {
//	mySubPub := KeygenRequestAddedSubPub(w.groupIDs)
//	myWatcher := &watcher.Watcher{
//		L:    w.L,
//		Subscribe: mySubPub.Subscribe,
//		Process:   mySubPub.Publish,
//		Filterer:  w.contractFilter,
//		P: w.P,
//	}
//	w.L.FatalOnError(myWatcher.Watch(ctx), "Failed to watch KeygenRequestAdded")
//	w.keygenRequestAddedWatcher = myWatcher
//}
//
//func (eh *Watcher) doWatchStakeRequestStarted(ctx context.Context) {
//	newSink := make(chan *contract.MpcManagerStakeRequestStarted)
//	err := eh.subscribeStakeRequestStarted(ctx, newSink, eh.pubKeyBytes)
//	if err == nil {
//		eh.sink = newSink
//		if eh.done != nil {
//			close(eh.done)
//		}
//		eh.done = make(chan struct{})
//		eh.watchStakeRequestStarted(ctx)
//	}
//}
//
//func (eh *Watcher) subscribeStakeRequestStarted(ctx context.Context, sink chan<- *contract.MpcManagerStakeRequestStarted, pubKeys [][]byte) (err error) {
//	if eh.sub != nil {
//		eh.sub.Unsubscribe()
//	}
//
//	err = backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
//		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//
//		newSub, err := filter.WatchStakeRequestStarted(nil, sink, pubKeys)
//		if err != nil {
//			return true, errors.WithStack(err)
//		}
//		eh.sub = newSub
//		return false, nil
//	})
//	err = errors.Wrapf(err, "failed to subscribe StakeRequestStarted event")
//	return
//}
//
//func (eh *Watcher) watchStakeRequestStarted(ctx context.Context) {
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-eh.done:
//				return
//			case evt := <-eh.sink:
//				evtObj := dispatcher.NewEvtObj(evt, nil)
//				eh.P.Publish(ctx, evtObj)
//				eh.L.Debug("StakeRequestStartedEvent emitted", []logger.Field{
//					{"StakeRequestStartedEvent", evt}}...)
//			case err := <-eh.sub.Err():
//				eh.L.ErrorOnError(err, "Got an error during watching StakeRequestStarted event")
//			}
//		}
//	}()
//}
//
//func (w *Watcher) EthClient() {
//
//}
