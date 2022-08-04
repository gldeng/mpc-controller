package watcher

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/contract/watcher"
	"github.com/avalido/mpc-controller/utils/dispatcher"
	"github.com/avalido/mpc-controller/utils/network/redialer"
	"github.com/avalido/mpc-controller/utils/network/redialer/adapter"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	"sync"
	"time"
)

type Watcher struct {
	Logger       logger.Logger
	PubKeys      [][]byte
	EthWsURL     string
	ContractAddr common.Address
	Publisher    dispatcher.Publisher

	contractFilterer bind.ContractFilterer
	ethClientCh      chan redialer.Client

	once sync.Once

	groupIDs [][32]byte

	watcherFactory *MpcManagerWatcherFactory

	participantAddedWatcher   *watcher.Watcher
	keygenRequestAddedWatcher *watcher.Watcher
	keyGeneratedWatcher       *watcher.Watcher
	stakeRequestAddedWatcher  *watcher.Watcher
	requestStartedWatcher     *watcher.Watcher
}

func (w *Watcher) Init(ctx context.Context) {
	w.once.Do(func() {
		reDialer := adapter.EthClientReDialer{
			Logger:        logger.Default(),
			EthURL:        w.EthWsURL,
			BackOffPolicy: backoff.ExponentialPolicy(0, time.Second, time.Second*10),
		}
		ethClient, ethclientCh, err := reDialer.GetClient(context.Background())
		w.Logger.FatalOnError(err, "Failed to get eth client")
		w.contractFilterer = ethClient.(*ethclient.Client)
		w.ethClientCh = ethclientCh

		boundFilterer, err := contract.BindMpcManagerFilterer(w.ContractAddr, w.contractFilterer)
		w.Logger.FatalOnError(err, "Failed to bind MpcManager filterer")

		watcherFactory := &MpcManagerWatcherFactory{w.Logger, boundFilterer}
		w.watcherFactory = watcherFactory
		err = w.watchParticipantAdded(ctx)
		w.Logger.FatalOnError(err, "Failed to watch ParticipantAdded")
	})
}

func (w *Watcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch _ := evtObj.Event.(type) {
	case *events.ParticipantAdded:
	case *events.KeygenRequestAdded:
	case *events.KeyGenerated:
	case *events.StakeRequestAdded:
	case *events.RequestStarted:
	}
}

// ParticipantAdded
func (w *Watcher) watchParticipantAdded(ctx context.Context) error {
	participantAddedWatcher, err := w.watcherFactory.NewWatcher(w.processParticipantAdded, nil, EvtParticipantAdded, []interface{}{w.groupIDs})
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtParticipantAdded)
	}
	w.participantAddedWatcher = participantAddedWatcher
	err = participantAddedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtParticipantAdded)
}

func (w *Watcher) processParticipantAdded(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerParticipantAdded)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.ParticipantAdded)(myEvt), nil))
	return nil
}

// KeygenRequestAdded
func (w *Watcher) watchKeygenRequestAdded(ctx context.Context) error {
	keygenRequestAddedWatcher, err := w.watcherFactory.NewWatcher(w.processKeygenRequestAdded, nil, EvtKeygenRequestAdded, []interface{}{w.groupIDs})
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtKeygenRequestAdded)
	}
	w.keygenRequestAddedWatcher = keygenRequestAddedWatcher
	err = keygenRequestAddedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtKeygenRequestAdded)
}

func (w *Watcher) processKeygenRequestAdded(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerKeygenRequestAdded)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.KeygenRequestAdded)(myEvt), nil))
	return nil
}

// KeyGenerated
func (w *Watcher) watchKeyGenerated(ctx context.Context) error {
	keyGeneratedWatcher, err := w.watcherFactory.NewWatcher(w.processKeyGenerated, nil, EvtKeyGenerated, []interface{}{w.groupIDs}) // todo: query
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtKeyGenerated)
	}
	w.keyGeneratedWatcher = keyGeneratedWatcher
	err = keyGeneratedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtKeyGenerated)
}

func (w *Watcher) processKeyGenerated(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerKeyGenerated)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.KeyGenerated)(myEvt), nil))
	return nil
}

// StakeRequestAdded
func (w *Watcher) watchStakeRequestAdded(ctx context.Context) error {
	stakeRequestAddedWatcher, err := w.watcherFactory.NewWatcher(w.processStakeRequestAdded, nil, EvtStakeRequestAdded, []interface{}{w.groupIDs}) // todo: query
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtStakeRequestAdded)
	}
	w.stakeRequestAddedWatcher = stakeRequestAddedWatcher
	err = stakeRequestAddedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtStakeRequestAdded)
}

func (w *Watcher) processStakeRequestAdded(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerStakeRequestAdded)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.StakeRequestAdded)(myEvt), nil))
	return nil
}

// RequestStarted
func (w *Watcher) watchRequestStarted(ctx context.Context) error {
	requestStartedWatcher, err := w.watcherFactory.NewWatcher(w.processRequestStarted, nil, EvtRequestStarted, []interface{}{w.groupIDs}) // todo: query
	if err != nil {
		return errors.Wrapf(err, "failed to create %v watcher", EvtRequestStarted)
	}
	w.requestStartedWatcher = requestStartedWatcher
	err = requestStartedWatcher.Watch(ctx)
	return errors.Wrapf(err, "failed to watch %v", EvtRequestStarted)
}

func (w *Watcher) processRequestStarted(ctx context.Context, evt interface{}) error { // todo: further process
	myEvt := evt.(*contract.MpcManagerRequestStarted)
	w.Publisher.Publish(ctx, dispatcher.NewEvtObj((*events.RequestStarted)(myEvt), nil))
	return nil
}
