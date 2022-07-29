package joining

import (
	"context"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// Accept event: *events.ContractFiltererCreatedEvent
// Accept event:  *events.GeneratedPubKeyInfoStoredEvent

// Emit event:  *contract.MpcManagerStakeRequestAdded

type StakeRequestAddedEventWatcher struct {
	Logger       logger.Logger
	ContractAddr common.Address
	Publisher    dispatcher.Publisher
	pubKeyBytes  [][]byte
	filterer     bind.ContractFilterer

	sub  event.Subscription
	sink chan *contract.MpcManagerStakeRequestAdded
	done chan struct{}

	StakeReqPublishDur time.Duration // server as rate limit
	StakeReqCacheCap   uint32
	stakeReqCacheQueue queue.Queue
	once               sync.Once
}

func (eh *StakeRequestAddedEventWatcher) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	eh.once.Do(func() {
		eh.stakeReqCacheQueue = queue.NewArrayQueue(int(eh.StakeReqCacheCap))
	})
	switch evt := evtObj.Event.(type) {
	case *events.ContractFiltererCreatedEvent:
		eh.filterer = evt.Filterer
	case *events.GeneratedPubKeyInfoStoredEvent:
		dnmPubKeyBtes, err := crypto.DenormalizePubKeyFromHex(evt.Val.CompressedGenPubKeyHex)
		if err != nil {
			eh.Logger.Error("Failed to denormalized generated public key", []logger.Field{{"error", err}}...)
			break
		}

		eh.pubKeyBytes = append(eh.pubKeyBytes, dnmPubKeyBtes)
	}
	if len(eh.pubKeyBytes) > 0 {
		eh.doWatchStakeRequestAdded(evtObj.Context)
	}
}

func (eh *StakeRequestAddedEventWatcher) doWatchStakeRequestAdded(ctx context.Context) {
	newSink := make(chan *contract.MpcManagerStakeRequestAdded, 1024)
	err := eh.subscribeStakeRequestAdded(ctx, newSink, eh.pubKeyBytes)
	if err == nil {
		eh.sink = newSink
		if eh.done != nil {
			close(eh.done)
		}
		eh.done = make(chan struct{})
		eh.watchStakeRequestAdded(ctx)
	}
}

func (eh *StakeRequestAddedEventWatcher) subscribeStakeRequestAdded(ctx context.Context, sink chan<- *contract.MpcManagerStakeRequestAdded, pubKeys [][]byte) error {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
	}

	err := backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (bool, error) {
		filter, err := contract.NewMpcManagerFilterer(eh.ContractAddr, eh.filterer)
		if err != nil {
			return true, errors.WithStack(err)
		}

		newSub, err := filter.WatchStakeRequestAdded(nil, sink, pubKeys)
		if err != nil {
			return true, errors.WithStack(err)
		}
		eh.sub = newSub
		return false, nil
	})
	err = errors.Wrapf(err, "failed to subscribe StakeRequestAdded event")
	return err
}

func (eh *StakeRequestAddedEventWatcher) watchStakeRequestAdded(ctx context.Context) {
	t := time.NewTicker(eh.StakeReqPublishDur)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-eh.done:
				return
			case evt := <-eh.sink:
				err := backoff.RetryFnConstantForever(ctx, time.Second, func() (retry bool, err error) {
					if err := eh.stakeReqCacheQueue.Enqueue(evt); err != nil {
						return true, errors.WithStack(err)
					}
					return false, nil
				})
				eh.Logger.ErrorOnError(err, "Failed to enqueue StakeRequestAdded event", []logger.Field{{"reqID", evt.RequestId}, {"txHash", evt.Raw.TxHash}}...)

				enqueued := eh.stakeReqCacheQueue.Count()
				capacity := eh.stakeReqCacheQueue.Capacity()
				eh.Logger.WarnOnTrue(float64(enqueued) > float64(capacity)*0.8,
					"Too many QUEUED StakeRequestAdded events",
					[]logger.Field{{"eventsQueued", enqueued}, {"queueCapacity", capacity}}...)
				eh.Logger.Debug("StakeRequestAdded event enqueued", []logger.Field{{"reqID", evt.RequestId}, {"txHash", evt.Raw.TxHash}, {"eventsEnqueued", enqueued}, {"queueCapacity", capacity}}...)
			case <-t.C:
				if !eh.stakeReqCacheQueue.Empty() {
					evt := eh.stakeReqCacheQueue.Dequeue().(*contract.MpcManagerStakeRequestAdded)
					evtObj := dispatcher.NewRootEventObject("StakeRequestAddedEventWatcher", evt, ctx)
					eh.Publisher.Publish(ctx, evtObj)
					eh.Logger.Debug("StakeRequestAdded event published", []logger.Field{{"reqID", evt.RequestId}, {"txHash", evt.Raw.TxHash}}...)
				}
			case err := <-eh.sub.Err():
				eh.Logger.ErrorOnError(err, "Got an error during watching StakeRequestAdded event")
			}
		}
	}()
}
