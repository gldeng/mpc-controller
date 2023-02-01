package subscriber

import (
	"encoding/json"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// ----------for log----------

var (
	enqueuedLogKey = []byte("enqueued-eth-log")
)

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

type enqueuedLog struct {
	BlockNumber uint64
	Index       uint
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

// ----------for stake request----------

var (
	stakeReqKey = []byte("stake-request")
)

type stakeReq struct {
	ReqNo          string
	LogBlockNumber uint64
	LogIndex       uint
}

func (s *Subscriber) foundStakeReq() (bool, error) {
	found, err := s.db.Exists(s.ctx, stakeReqKey)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return found, nil
}

func (s *Subscriber) loadStakeReq() (*stakeReq, error) {
	storedBytes, err := s.db.Get(s.ctx, stakeReqKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var result stakeReq
	_ = json.Unmarshal(storedBytes, &result)
	return &result, nil
}

func (s *Subscriber) saveStakeReq(req *stakeReq) error {
	bytes, _ := json.Marshal(*req)
	err := s.db.Set(s.ctx, stakeReqKey, bytes)
	return errors.WithStack(err)
}
