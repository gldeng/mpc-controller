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
	"time"
)

type Subscriber struct {
	ctx           context.Context
	logger        logger.Logger
	config        core.Config
	client        *ethclient.Client
	subscription  ethereum.Subscription
	eventLogQueue Queue
	eventIDGetter EventIDGetter
	filter        ethereum.FilterQuery
	resubscribeCh chan struct{}
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
		filter:        ethereum.FilterQuery{Addresses: []common.Address{config.MpcManagerAddress}},
		resubscribeCh: make(chan struct{}),
	}, nil
}

func (s *Subscriber) Start() error {
	client, err := s.config.CreateWsClient()
	if err != nil {
		return errors.Wrap(err, "failed to create ws client")
	}
	s.client = client

	if err := s.subscribe(); err != nil {
		return errors.Wrap(err, "failed to subscribe")
	}

	go s.resubscribe()
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

func (s *Subscriber) subscribe() error {
	eventLogs := make(chan types.Log, 1024)
	sub, err := s.client.SubscribeFilterLogs(s.ctx, s.filter, eventLogs)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to contract events")
	}

	s.subscription = event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-eventLogs:
				s.updateMetrics(log)
				if err := s.eventLogQueue.Enqueue(log); err != nil {
					s.logger.Error("failed to enqueue event log", []logger.Field{{"error", err}}...)
				}
			case err := <-sub.Err():
				s.logger.Error("got an error for subscription", []logger.Field{{"error", err}}...)
				s.resubscribeCh <- struct{}{}
				return err
			case <-quit:
				return nil
			}
		}
	})
	return nil
}

func (s *Subscriber) resubscribe() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.resubscribeCh:
			s.subscription.Unsubscribe()
			s.client.Close()

		resubscribe:
			for {
				select {
				case <-s.ctx.Done():
					return
				default:
					client, err := s.config.CreateWsClient()
					if err != nil {
						s.logger.Error("failed to recreate ws client", []logger.Field{{"error", err}}...)
						break
					}
					s.logger.Debug("ws client recreated")
					s.client = client

					if err := s.subscribe(); err != nil {
						s.logger.Error("failed to re-subscribe", []logger.Field{{"error", err}}...)
						break
					}
					s.logger.Debug("re-subscribed")
					break resubscribe
				}
				time.Sleep(time.Second)
			}
		}
	}
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
