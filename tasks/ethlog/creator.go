package ethlog

import (
	binding "github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/tasks/keygen"
	"github.com/avalido/mpc-controller/tasks/recovery"
	"github.com/avalido/mpc-controller/tasks/stake"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"math/big"
)

var (
	_ core.LogEventHandler = &RequestCreator{}
)

type RequestCreator struct {
	lastStakeReqNo *big.Int
}

func (c *RequestCreator) Handle(ctx core.EventHandlerContext, log types.Log) ([]core.Task, error) {

	if log.Topics[0] == ctx.GetEventID(core.EvtRequestStarted) {
		// Ignore
		return nil, nil
		/*
			prom.ContractEvtRequestStarted.Inc()
			event := new(binding.MpcManagerRequestStarted)
			err := ctx.GetContract().UnpackLog(event, EvtRequestStarted, log)
			if err != nil {
				return nil, errors.Wrap(err, "failed to unpack log")
			}
			event.Raw = log
			task := NewRequestStartedHandler(*event)
			return []core.Task{task}, nil
		*/
	}
	if log.Topics[0] == ctx.GetEventID(core.EvtParticipantAdded) {
		event := new(binding.MpcManagerParticipantAdded)
		err := ctx.GetContract().UnpackLog(event, core.EvtParticipantAdded, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unpack log")
		}
		event.Raw = log
		task := NewParticipantAddedHandler(*event)
		return []core.Task{task}, nil
	}
	if log.Topics[0] == ctx.GetEventID(core.EvtKeygenRequestAdded) {
		event := new(binding.MpcManagerKeygenRequestAdded)
		err := ctx.GetContract().UnpackLog(event, core.EvtKeygenRequestAdded, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unpack log")
		}
		event.Raw = log
		task := keygen.NewRequestAdded(*event)
		return []core.Task{task}, nil
	}
	if log.Topics[0] == ctx.GetEventID(core.EvtKeyGenerated) {
		event := new(binding.MpcManagerKeyGenerated)
		err := ctx.GetContract().UnpackLog(event, core.EvtKeyGenerated, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unpack log")
		}
		event.Raw = log
		task := NewKeyGeneratedHandler(*event)
		return []core.Task{task}, nil
	}
	if log.Topics[0] == ctx.GetEventID(core.EvtStakeRequestAdded) {
		event := new(binding.MpcManagerStakeRequestAdded)
		err := ctx.GetContract().UnpackLog(event, core.EvtStakeRequestAdded, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unpack log")
		}
		c.checkStakeReqNoContinuity(ctx, event.RequestNumber, log.BlockNumber, log.TxHash)
		task, err := stake.NewStakeJoinAndStake(*event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create task")
		}
		return []core.Task{task}, nil
	}
	if log.Topics[0] == ctx.GetEventID(core.EvtRequestFailed) {
		event := new(binding.MpcManagerRequestFailed)
		err := ctx.GetContract().UnpackLog(event, core.EvtRequestFailed, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unpack log")
		}
		task, err := recovery.NewJoinAndRecover(*event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create task")
		}
		return []core.Task{task}, nil
	}
	return nil, nil
}

func (c *RequestCreator) checkStakeReqNoContinuity(ctx core.EventHandlerContext, newStakeReqNo *big.Int, blockNumber uint64, txHash common.Hash) {
	defer func() {
		c.lastStakeReqNo = newStakeReqNo
	}()

	if c.lastStakeReqNo != nil {
		one := big.NewInt(1)
		plusOne := new(big.Int).Add(c.lastStakeReqNo, one)
		if plusOne.Cmp(newStakeReqNo) != 0 {
			prom.DiscontinuousValue.With(prometheus.Labels{"checker": "ethlog", "field": "stakeReqNo"}).Inc()
			ctx.GetLogger().Warn("got discontinuous stake request number", []logger.Field{
				{"expectedStakeReqNo", plusOne},
				{"gotStakeReqNo", newStakeReqNo},
				{"blockNumber", blockNumber},
				{"txHash", txHash}}...)
		}
	}
}
