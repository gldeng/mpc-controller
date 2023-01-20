package subscriber

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
)

// compensateMissedLogs compensate occasionally missed eth log because of bad situations, such as network partition
func (s *Subscriber) compensateMissedLogs() error {
	s.logger.Debug("prepare to compensate missed logs")

	found, err := s.foundEnqueuedLog()
	if err != nil {
		prom.EventCompensationError.With(prometheus.Labels{"type": "ethLog", "reason": "readDb"}).Inc()
		return errors.Wrap(err, "failed to check the existence of enqueued log")
	}

	if !found {
		s.logger.Debug("no need to compensate: no enqueued log found")
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

	missedLogs := s.filterMissedLogs(maybeMissedLogs, storedLog)
	if len(missedLogs) == 0 {
		s.logger.Debug("no need to compensate: no missed logs found")
		return nil
	}

	s.logger.Debug("continue to compensate missed logs")
	err = s.enqueueAndSaveLogs(missedLogs)
	if err != nil {
		prom.EventCompensationError.With(prometheus.Labels{"type": "ethLog", "reason": "enqueueAndSaveLog"}).Inc()
		return errors.Wrap(err, "failed to enqueue and save logs")
	}
	return nil
}

func (s *Subscriber) filterMissedLogs(maybeMissedLogs []types.Log, storedLog enqueuedLog) []types.Log {
	var missedLogs []types.Log
	for _, maybeMissedLog := range maybeMissedLogs {
		if maybeMissedLog.Removed {
			prom.EventReverted.With(prometheus.Labels{"type": "ethLog"}).Inc()
			s.logger.Warn("invalid log to compensate: log reverted", []logger.Field{
				{"blockNumber", maybeMissedLog.BlockNumber},
				{"index", maybeMissedLog.Index}}...)
			continue
		}
		if maybeMissedLog.BlockNumber < storedLog.BlockNumber {
			s.logger.Warn("invalid log to compensate: block number unaccepted", []logger.Field{
				{"blockNumber", maybeMissedLog.BlockNumber},
				{"index", maybeMissedLog.Index}}...)
			continue
		}
		if maybeMissedLog.BlockNumber == storedLog.BlockNumber {
			if maybeMissedLog.Index <= storedLog.Index {
				s.logger.Warn("invalid log to compensate: index unaccepted", []logger.Field{
					{"blockNumber", maybeMissedLog.BlockNumber},
					{"index", maybeMissedLog.Index}}...)
				continue
			}
			s.logger.Debug("found missed log", []logger.Field{
				{"blockNumber", maybeMissedLog.BlockNumber},
				{"index", maybeMissedLog.Index}}...)
			missedLogs = append(missedLogs, maybeMissedLog)
			continue
		}
		s.logger.Debug("found missed log", []logger.Field{
			{"blockNumber", maybeMissedLog.BlockNumber},
			{"index", maybeMissedLog.Index}}...)
		missedLogs = append(missedLogs, maybeMissedLog)
	}
	if len(missedLogs) == 0 {
		s.logger.Debug("no missed log found")
		return nil
	}
	return missedLogs
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
		prom.EventCompensation.With(prometheus.Labels{"type": "ethLog"}).Inc()
	}
	return nil
}
