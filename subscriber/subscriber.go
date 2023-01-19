package subscriber

import (
	"context"
	"encoding/json"
	binding "github.com/avalido/mpc-controller/contract"
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
	"math/big"
	"time"
)

// TODO: make it configurable?
const (
	eventLogChanCapacity    = 1024
	queueBufferChanCapacity = 2048
)

var (
	enqueuedLogKey = []byte("enqueued-eth-log")
)

type Subscriber struct {
	ctx            context.Context
	logger         logger.Logger
	config         core.Config
	client         *ethclient.Client
	subscription   ethereum.Subscription
	eventLogQueue  Queue
	eventIDGetter  EventIDGetter
	filter         ethereum.FilterQuery
	backoffMax     time.Duration
	lastStakeReqNo *big.Int
	logUnpacker    LogUnpacker
	queueBufferCh  chan types.Log
	db             core.Store
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
		queueBufferCh: make(chan types.Log, queueBufferChanCapacity),
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

		eventLogs := make(chan types.Log, eventLogChanCapacity)
		sub, err := s.client.SubscribeFilterLogs(s.ctx, s.filter, eventLogs)
		if err != nil {
			prom.ContractEvtSubErr.Inc()
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
					s.updateMetricOnEvtID(log)
					s.logger.Debug("received event log", []logger.Field{
						{"address", log.Address},
						{"txHash", log.TxHash},
						{"topics", log.Topics},
						{"blockNumber", log.BlockNumber},
						{"index", log.Index},
						{"removed", log.Removed}}...)

					valid, err := s.verifyReceivedLog(log)
					if err != nil {
						s.logger.Error("failed to verify received log", []logger.Field{
							{"blockNumber", log.BlockNumber},
							{"index", log.Index},
							{"error", err}}...)
						continue
					}

					if !valid {
						continue
					}

					s.checkStakeReqNoContinuity(log)

					s.queueBufferCh <- log
					if err := s.eventLogQueue.Enqueue(<-s.queueBufferCh); err != nil {
						prom.QueueOperationError.With(prometheus.Labels{"pkg": "subscriber", "operation": "enqueue"}).Inc()
						s.logger.Error("failed to enqueue event log, enqueue error", []logger.Field{
							{"blockNumber", log.BlockNumber},
							{"index", log.Index},
							{"error", err}}...)
						continue
					}
					prom.QueueOperation.With(prometheus.Labels{"pkg": "subscriber", "operation": "enqueue"}).Inc()
					s.logger.Debug("enqueued event log", []logger.Field{
						{"blockNumber", log.BlockNumber},
						{"index", log.Index}}...)

					err = s.saveEnqueuedLog(enqueuedLog{log.BlockNumber, log.Index})
					if err != nil {
						s.logger.Error("failed to save enqueued event log", []logger.Field{
							{"blockNumber", log.BlockNumber},
							{"index", log.Index},
							{"error", err}}...)
						prom.DBOperationError.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
						continue
					}
					prom.DBOperation.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
					s.logger.Debug("saved enqueued event log", []logger.Field{
						{"blockNumber", log.BlockNumber},
						{"index", log.Index}}...)
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

type enqueuedLog struct {
	BlockNumber uint64
	Index       uint
}

// validateReceivedLog verifies the block number and index of a received eth log against those of stored enqueued log
func (s *Subscriber) verifyReceivedLog(receivedLog types.Log) (bool, error) {
	found, err := s.db.Exists(s.ctx, enqueuedLogKey)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if !found {
		return true, nil
	}

	storedBytes, err := s.db.Get(s.ctx, enqueuedLogKey)
	if err != nil {
		return false, errors.WithStack(err)
	}

	var storedLog enqueuedLog
	_ = json.Unmarshal(storedBytes, &storedLog)

	if receivedLog.BlockNumber < storedLog.BlockNumber {
		prom.InvalidReceivedEvent.With(prometheus.Labels{"type": "ethLog", "field": "blockNumber"}).Inc()
		s.logger.Error("invalid block number", []logger.Field{
			{"minimumAccepted", storedLog.BlockNumber},
			{"got", receivedLog.BlockNumber}}...)
		return false, nil
	}

	if receivedLog.BlockNumber == storedLog.BlockNumber && receivedLog.Index <= storedLog.Index {
		prom.InvalidReceivedEvent.With(prometheus.Labels{"type": "ethLog", "field": "index"}).Inc()
		s.logger.Error("invalid block number", []logger.Field{
			{"minimumAccepted", storedLog.Index + 1},
			{"got", receivedLog.Index}}...)
		return false, nil
	}

	return true, nil
}

func (s *Subscriber) saveEnqueuedLog(log enqueuedLog) error {
	logBytes, _ := json.Marshal(log)
	err := s.db.Set(s.ctx, enqueuedLogKey, logBytes)
	return errors.WithStack(err)
}

func (s *Subscriber) updateMetricOnEvtID(log types.Log) {
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

func (s *Subscriber) checkStakeReqNoContinuity(log types.Log) {
	switch log.Topics[0] {
	case s.eventIDGetter.GetEventID(core.EvtStakeRequestAdded):
		event := new(binding.MpcManagerStakeRequestAdded)
		err := s.logUnpacker.UnpackLog(event, core.EvtStakeRequestAdded, log)
		if err != nil {
			s.logger.Error("failed to unpack log for EvtStakeRequestAdded", []logger.Field{{"error", err}}...)
		}

		newStakeReqNo := event.RequestNumber

		defer func() {
			s.lastStakeReqNo = newStakeReqNo
		}()

		if s.lastStakeReqNo != nil {
			one := big.NewInt(1)
			plusOne := new(big.Int).Add(s.lastStakeReqNo, one)
			if plusOne.Cmp(newStakeReqNo) != 0 {
				prom.DiscontinuousValue.With(prometheus.Labels{"checker": "subscriber", "field": "stakeReqNo"}).Inc()
				s.logger.Warn("got discontinuous stake request number", []logger.Field{
					{"expectedStakeReqNo", plusOne},
					{"gotStakeReqNo", newStakeReqNo},
					{"blockNumber", log.BlockNumber},
					{"txHash", log.TxHash}}...)
			}
		}
	}
}
