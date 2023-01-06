package ethlog

import (
	binding "github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/tasks/keygen"
	"github.com/avalido/mpc-controller/tasks/stake"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
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
		c.checkStakeReqNoContinuity(ctx, event.RequestNumber)
		task, err := stake.NewStakeJoinAndStake(*event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create task")
		}
		return []core.Task{task}, nil
	}
	return nil, nil
}

func (c *RequestCreator) checkStakeReqNoContinuity(ctx core.EventHandlerContext, newStakeReqNo *big.Int) {
	defer func() {
		c.lastStakeReqNo = newStakeReqNo
	}()

	if c.lastStakeReqNo != nil {
		one := big.NewInt(1)
		plusOne := new(big.Int).Add(c.lastStakeReqNo, one)
		if plusOne.Cmp(newStakeReqNo) != 0 {
			prom.InconsistentStakeReqNo.Inc()
			ctx.GetLogger().Warn("got inconsistent stake request number", []logger.Field{
				{"expectedStakeReqNo", plusOne},
				{"gotStakeReqNo", newStakeReqNo}}...)
		}
	}
}
