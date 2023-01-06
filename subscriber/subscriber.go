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
	backoffMax    time.Duration
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
		backoffMax:    time.Second * 5,
	}, nil
}

func (s *Subscriber) Start() error {
	client, err := s.config.CreateWsClient()
	if err != nil {
		return errors.Wrap(err, "failed to create ws client")
	}
	s.client = client

	resubscribeErrFunc := func(_ context.Context, err error) (event.Subscription, error) {
		if err != nil {
			s.logger.Error("got an error for subscription", []logger.Field{{"error", err}}...)
		}

		eventLogs := make(chan types.Log, 1024)
		sub, err := s.client.SubscribeFilterLogs(s.ctx, s.filter, eventLogs)
		if err != nil {
			s.logger.Error("failed to subscribe contract events", []logger.Field{{"error", err}}...)
			return nil, err
		}
		s.logger.Debug("subscribed contract events")

		go func() {
			for {
				select {
				case <-s.ctx.Done():
					return
				case log := <-eventLogs:
					s.updateMetrics(log)
					if err := s.eventLogQueue.Enqueue(log); err != nil {
						s.logger.Error("failed to enqueue event log", []logger.Field{{"error", err}}...)
					}
				case <-sub.Err():
					return
				}
			}
		}()

		return sub, nil
	}

	s.subscription = event.ResubscribeErr(s.backoffMax, resubscribeErrFunc)
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
