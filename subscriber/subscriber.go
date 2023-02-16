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
	"github.com/prometheus/client_golang/prometheus"
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
	logUnpacker   LogUnpacker
	queueBufferCh chan types.Log
	db            core.Store
}

type Queue interface {
	Enqueue(value interface{}) error
}

type EventIDGetter interface {
	GetEventID(event string) common.Hash
}

type LogUnpacker interface {
	UnpackLog(out interface{}, event string, log types.Log) error
}

func NewSubscriber(ctx context.Context, logger logger.Logger, config core.Config, eventLogQueue Queue, evtIDGetter EventIDGetter, unpacker LogUnpacker, db core.Store) (*Subscriber, error) {
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
		logUnpacker:   unpacker,
		queueBufferCh: make(chan types.Log, core.DefaultParameters.QueueBufferChanCapacity),
		db:            db,
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

		err = s.compensateMissedLogs()
		if err != nil {
			s.logger.Error("failed to compensate missed logs", []logger.Field{{"error", err}}...)
			return nil, err
		}

		eventLogs := make(chan types.Log, core.DefaultParameters.EventLogChanCapacity)
		sub, err := s.client.SubscribeFilterLogs(s.ctx, s.filter, eventLogs)
		if err != nil {
			prom.ContractEvtSubErr.Inc()
			s.logger.Error("failed to subscribe contract events", []logger.Field{{"error", err}}...)
			return nil, err
		}
		prom.ContractEvtSub.Inc()
		s.logger.Debug("subscribed contract events")

		go func() {
			for {
				select {
				case <-s.ctx.Done():
					return
				case log := <-eventLogs:
					s.updateMetricOnEvtID(log)

					err := s.verifyStreamingLog(log)
					if err != nil {
						s.logger.Error("failed to verify streaming log", []logger.Field{
							{"blockNumber", log.BlockNumber},
							{"index", log.Index},
							{"error", err}}...)
						continue
					}

					err = s.checkStakeReqNoContinuity(log)
					if err != nil {
						s.logger.Error("failed to check the continuity of stake request number")
					}
					_ = s.enqueueAndSaveLog(log)
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

// verifyStreamingLog verifies the block number and index of a streaming eth log against those of stored enqueued log
func (s *Subscriber) verifyStreamingLog(streamingLog types.Log) error {
	found, err := s.foundEnqueuedLog()
	if err != nil {
		return errors.Wrap(err, "failed to check enqueued log existence")
	}

	if !found {
		return nil
	}

	storedLog, err := s.loadEnqueuedLog()
	if err != nil {
		return errors.Wrap(err, "failed to load enqueued log")
	}

	if streamingLog.BlockNumber < storedLog.BlockNumber {
		prom.InvalidStreamingEvent.With(prometheus.Labels{"type": "ethLog", "field": "blockNumber"}).Inc()
		return errors.Errorf("invalid streaming log: block number unaccepted, minimumAccepted:%v, got:%v", storedLog.BlockNumber, streamingLog.BlockNumber)
	}

	if streamingLog.BlockNumber == storedLog.BlockNumber && streamingLog.Index <= storedLog.Index {
		prom.InvalidStreamingEvent.With(prometheus.Labels{"type": "ethLog", "field": "index"}).Inc()
		return errors.Errorf("invalid streaming log: index unaccepted, minimumAccepted:%v, got:%v", storedLog.Index+1, streamingLog.Index)
	}

	return nil
}

func (s *Subscriber) updateMetricOnEvtID(log types.Log) {
	s.logger.Debug("received event log", []logger.Field{
		{"address", log.Address},
		{"txHash", log.TxHash},
		{"topics", log.Topics},
		{"blockNumber", log.BlockNumber},
		{"index", log.Index},
		{"removed", log.Removed}}...)

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
