package subscriber

import (
	binding "github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
)

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
