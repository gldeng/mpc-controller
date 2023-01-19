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

		err = s.compensateMissedLogs()
		if err != nil {
			s.logger.Error("failed to compensate missed logs", []logger.Field{{"error", err}}...)
			return nil, err
		}
		prom.EventCompensation.With(prometheus.Labels{"type": "ethLog"}).Inc()

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

					err := s.verifyStreamingLog(log)
					if err != nil {
						s.logger.Error("failed to verify streaming log", []logger.Field{
							{"blockNumber", log.BlockNumber},
							{"index", log.Index},
							{"error", err}}...)
						continue
					}

					s.checkStakeReqNoContinuity(log)
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

type enqueuedLog struct {
	BlockNumber uint64
	Index       uint
}

// compensateMissedLogs compensate missed eth log because of exceptions, such as network partition
func (s *Subscriber) compensateMissedLogs() error {
	found, err := s.foundEnqueuedLog()
	if err != nil {
		prom.EventCompensationError.With(prometheus.Labels{"type": "ethLog", "reason": "readDb"}).Inc()
		return errors.Wrap(err, "failed to check the existence of enqueued log")
	}

	if !found {
		return nil
	}

	storedLog, err := s.loadEnqueuedLog()
	if err != nil {
		prom.EventCompensationError.With(prometheus.Labels{"type": "ethLog", "reason": "readDb"}).Inc()
		return errors.Wrap(err, "failed to load enqueued log")
	}

	maybeMissedLogs, err := s.filterLatestLogs(storedLog.BlockNumber)
	if err != nil {
		prom.EventCompensationError.With(prometheus.Labels{"type": "ethLog", "reason": "filterLog"}).Inc()
		return errors.Wrap(err, "failed to filter latest logs")
	}

	if len(maybeMissedLogs) == 0 {
		return nil
	}

	for _, maybeMissedLog := range maybeMissedLogs {
		s.logger.Debug("maybe missed log", []logger.Field{
			{"blockNumber", maybeMissedLog.BlockNumber},
			{"index", maybeMissedLog.Index},
			{}}...)
	}

	var missedFrom int

	for i, maybeMissedLog := range maybeMissedLogs {
		if maybeMissedLog.BlockNumber < storedLog.BlockNumber {
			continue
		}
		if maybeMissedLog.BlockNumber == storedLog.BlockNumber {
			if maybeMissedLog.Index <= storedLog.Index {
				continue
			}
			missedFrom = i
			break
		}
		missedFrom = i
		break
	}

	missedLogs := maybeMissedLogs[missedFrom:]
	if len(missedLogs) == 0 {
		s.logger.Debug("no missed log found")
		return nil
	}

	for _, missedLog := range maybeMissedLogs {
		s.logger.Debug("found missed log", []logger.Field{
			{"blockNumber", missedLog.BlockNumber},
			{"index", missedLog.Index}}...)
	}

	err = s.enqueueAndSaveLogs(missedLogs)
	if err != nil {
		prom.EventCompensationError.With(prometheus.Labels{"type": "ethLog", "reason": "enqueueAndSaveLog"}).Inc()
		return errors.Wrap(err, "failed to enqueue and save logs")
	}
	return nil
}

func (s *Subscriber) filterLatestLogs(fromBlock uint64) ([]types.Log, error) {
	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{s.config.MpcManagerAddress},
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   nil, // latest block
		Topics:    nil, // any topic list
	}
	logs, err := s.client.FilterLogs(s.ctx, filterQuery)
	return logs, errors.Wrap(err, "failed to filter logs")
}

func (s *Subscriber) enqueueAndSaveLogs(logs []types.Log) error {
	for _, log := range logs {
		s.updateMetricOnEvtID(log)
		s.checkStakeReqNoContinuity(log)

		err := s.enqueueAndSaveLog(log)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (s *Subscriber) enqueueAndSaveLog(log types.Log) error {
	var gotErr error
	defer func() {
		if gotErr != nil {
			s.logger.Error("failed to enqueue and save event log", []logger.Field{
				{"blockNumber", log.BlockNumber},
				{"index", log.Index},
				{"error", gotErr}}...)
			return
		}
		s.logger.Debug("enqueued and saved event log", []logger.Field{
			{"blockNumber", log.BlockNumber},
			{"index", log.Index}}...)
	}()

	s.queueBufferCh <- log
	if err := s.eventLogQueue.Enqueue(<-s.queueBufferCh); err != nil {
		prom.QueueOperationError.With(prometheus.Labels{"pkg": "subscriber", "operation": "enqueue"}).Inc()
		gotErr = err
		return errors.WithStack(err)
	}
	prom.QueueOperation.With(prometheus.Labels{"pkg": "subscriber", "operation": "enqueue"}).Inc()

	err := s.saveEnqueuedLog(enqueuedLog{log.BlockNumber, log.Index})
	if err != nil {
		prom.DBOperationError.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
		gotErr = err
		return errors.WithStack(err)
	}
	prom.DBOperation.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
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
		return errors.Errorf("invalid block number, minimumAccepted:%v, got:%v", storedLog.BlockNumber, streamingLog.BlockNumber)
	}

	if streamingLog.BlockNumber == storedLog.BlockNumber && streamingLog.Index <= storedLog.Index {
		prom.InvalidStreamingEvent.With(prometheus.Labels{"type": "ethLog", "field": "index"}).Inc()
		return errors.Errorf("invalid index, minimumAccepted:%v, got:%v", storedLog.Index+1, streamingLog.Index)
	}

	return nil
}

func (s *Subscriber) foundEnqueuedLog() (bool, error) {
	found, err := s.db.Exists(s.ctx, enqueuedLogKey)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return found, nil
}

func (s *Subscriber) loadEnqueuedLog() (enqueuedLog, error) {
	storedBytes, err := s.db.Get(s.ctx, enqueuedLogKey)
	if err != nil {
		return enqueuedLog{}, errors.WithStack(err)
	}

	var storedLog enqueuedLog
	_ = json.Unmarshal(storedBytes, &storedLog)
	return storedLog, nil
}

func (s *Subscriber) saveEnqueuedLog(log enqueuedLog) error {
	logBytes, _ := json.Marshal(log)
	err := s.db.Set(s.ctx, enqueuedLogKey, logBytes)
	return errors.WithStack(err)
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

func (s *Subscriber) checkStakeReqNoContinuity(log types.Log) {
	switch log.Topics[0] {
	case s.eventIDGetter.GetEventID(core.EvtStakeRequestAdded):
		event := new(binding.MpcManagerStakeRequestAdded)
		err := s.logUnpacker.UnpackLog(event, core.EvtStakeRequestAdded, log)
		if err != nil {
			s.logger.Error("failed to unpack log for EvtStakeRequestAdded", []logger.Field{{"error", err}}...)
			return
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
