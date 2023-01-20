package subscriber

import (
	binding "github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
)

func (s *Subscriber) checkStakeReqNoContinuity(log types.Log) error {
	switch log.Topics[0] {
	case s.eventIDGetter.GetEventID(core.EvtStakeRequestAdded):
		event := new(binding.MpcManagerStakeRequestAdded)
		err := s.logUnpacker.UnpackLog(event, core.EvtStakeRequestAdded, log)
		if err != nil {
			return errors.Wrap(err, "failed to unpack log")
		}

		newStakeReqNo := event.RequestNumber

		found, err := s.foundStakeReqNo()
		if err != nil {
			return errors.Wrap(err, "failed to check the existence of stake request number")
		}

		if !found {
			s.logger.Debug("stake request number not found")
			err := s.saveStakeReqNo(stakeReqNo(newStakeReqNo.String()))
			if err != nil {
				prom.DBOperationError.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
				return errors.Wrapf(err, "failed to insert stake request number %v", newStakeReqNo.String())
			}
			prom.DBOperation.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
			s.logger.Debug("inserted stake request number", []logger.Field{{"stakeReqNo", newStakeReqNo.String()}}...)
			return nil
		}

		storeStakeReqNo, err := s.loadStakeReqNo()
		if err != nil {
			prom.DBOperationError.With(prometheus.Labels{"pkg": "subscriber", "operation": "load"}).Inc()
			return errors.Wrap(err, "failed to load stake request number")
		}
		prom.DBOperation.With(prometheus.Labels{"pkg": "subscriber", "operation": "load"}).Inc()
		s.logger.Debug("loaded stake request number", []logger.Field{{"stakeReqNo", storeStakeReqNo}}...)

		lastStakeReqNo, _ := new(big.Int).SetString(string(storeStakeReqNo), 10)
		one := big.NewInt(1)
		plusOne := new(big.Int).Add(lastStakeReqNo, one)
		if plusOne.Cmp(newStakeReqNo) != 0 {
			prom.DiscontinuousValue.With(prometheus.Labels{"checker": "subscriber", "field": "stakeReqNo"}).Inc()
			s.logger.Warn("got discontinuous stake request number", []logger.Field{
				{"expectedStakeReqNo", plusOne},
				{"gotStakeReqNo", newStakeReqNo},
				{"blockNumber", log.BlockNumber},
				{"txHash", log.TxHash}}...)
		}
		err = s.saveStakeReqNo(stakeReqNo(newStakeReqNo.String()))
		if err != nil {
			prom.DBOperationError.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
			return errors.Wrapf(err, "failed to update stake request number %v", newStakeReqNo.String())
		}
		prom.DBOperation.With(prometheus.Labels{"pkg": "subscriber", "operation": "save"}).Inc()
		s.logger.Debug("updated stake request number", []logger.Field{{"stakeReqNo", newStakeReqNo.String()}}...)
		return nil
	default:
		return nil
	}
}
