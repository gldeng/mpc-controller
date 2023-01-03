package subscriber

import (
	"context"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type Subscriber struct {
	ctx            context.Context
	logger         logger.Logger
	config         core.Config
	client         *ethclient.Client
	subscription   ethereum.Subscription
	eventLogQueue  Queue
	eventIDGetter  EventIDGetter
	clientRenewCh  chan struct{}
	tryToReconnect bool
	once           sync.Once
}

type Queue interface {
	Enqueue(value interface{}) error
}

type EventIDGetter interface {
	GetEventID(event string) common.Hash
}

func NewSubscriber(ctx context.Context, logger logger.Logger, config core.Config, eventLogQueue Queue, evtIDGetter EventIDGetter) (*Subscriber, error) {
	return &Subscriber{
		ctx:           ctx,
		logger:        logger,
		config:        config,
		client:        nil,
		subscription:  nil,
		eventLogQueue: eventLogQueue,
		eventIDGetter: evtIDGetter,
		clientRenewCh: make(chan struct{}),
	}, nil
}

func (s *Subscriber) Start() error {
	if s.client == nil {
		client, err := s.config.CreateWsClient()
		if err != nil {
			return errors.Wrap(err, "failed to create ws client")
		}
		s.client = client
	}

	filter := ethereum.FilterQuery{
		Addresses: []common.Address{s.config.MpcManagerAddress},
	}

	eventLogs := make(chan types.Log, 1024)
	sub, err := s.client.SubscribeFilterLogs(s.ctx, filter, eventLogs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to contract events")
	}

	s.subscription = event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-eventLogs:
				s.updateMetrics(log)
				s.eventLogQueue.Enqueue(log) // TODO: Handle failure. There should be only one publisher to this queue, so now it should be fine to ignore the error.
			case err := <-sub.Err():
				// TODO: Should we reconnect if this happens?
				return err
			case <-quit:
				return nil
			}
		}
	})

	s.once.Do(func() {
		go s.restartOnWsClientRecreated()
		go s.recreateWsClientOnConnectionFailure()
	})

	return nil
}

func (s *Subscriber) Close() error {
	if s.subscription != nil {
		s.subscription.Unsubscribe()
		s.subscription = nil
	}
	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
	return nil
}

func (s *Subscriber) updateMetrics(log types.Log) {
	switch log.Topics[0] {
	case s.eventIDGetter.GetEventID(core.EvtRequestStarted):
	case s.eventIDGetter.GetEventID(core.EvtParticipantAdded):
		prom.ContractEvtParticipantAdded.Inc()
	case s.eventIDGetter.GetEventID(core.EvtKeygenRequestAdded):
		prom.ContractEvtKeygenRequestAdded.Inc()
	case s.eventIDGetter.GetEventID(core.EvtKeyGenerated):
		prom.ContractEvtKeyGenerated.Inc()
	case s.eventIDGetter.GetEventID(core.EvtStakeRequestAdded):
		prom.ContractEvtStakeRequestAdded.Inc()
	}
}

func (s *Subscriber) restartOnWsClientRecreated() {
	s.logger.Debug("waiting for ws client recreated")

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.clientRenewCh:
			for {
				if _, ok := <-s.ctx.Done(); ok {
					return
				}

				if err := s.Close(); err != nil {
					s.logger.Error("failed to close Subscriber", []logger.Field{{"error", err}}...)
					time.Sleep(time.Second)
					continue
				}
				if err := s.Start(); err != nil {
					s.logger.Error("failed to restart Subscriber", []logger.Field{{"error", err}}...)
					time.Sleep(time.Second)
					continue
				}

				s.logger.Debug("restarted Subscriber")
			}
		}
	}
}

func (s *Subscriber) recreateWsClientOnConnectionFailure() {
	s.logger.Debug("checking ws connectivity")

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
			if !s.tryToReconnect {
				_, err := s.client.NetworkID(s.ctx)
				if err != nil {
					s.logger.Error("failed to check ws connection", []logger.Field{{"error", err}}...)
					s.tryToReconnect = true
				}
			}

			if s.tryToReconnect {
				client, err := s.config.CreateWsClient()
				if err != nil {
					s.logger.Error("failed to recreate ws client", []logger.Field{{"error", err}}...)
					continue
				}
				s.client = client
				s.clientRenewCh <- struct{}{}
				s.logger.Debug("ws client recreated")
				s.tryToReconnect = false
			}
		}
	}
}
