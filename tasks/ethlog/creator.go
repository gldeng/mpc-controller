package ethlog

import (
	binding "github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

var (
	_ core.LogEventHandler = &RequestCreator{}
)

type RequestCreator struct {
}

func (c *RequestCreator) Handle(ctx core.EventHandlerContext, log types.Log) ([]core.Task, error) {

	if log.Topics[0] == ctx.GetEventID(EvtRequestStarted) {
		event := new(binding.MpcManagerRequestStarted)
		err := ctx.GetContract().UnpackLog(event, EvtRequestStarted, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unpack log")
		}
		task := NewRequestStartedHandler(*event)
		return []core.Task{task}, nil
	}
	if log.Topics[0] == ctx.GetEventID(EvtParticipantAdded) {
		event := new(binding.MpcManagerParticipantAdded)
		err := ctx.GetContract().UnpackLog(event, EvtParticipantAdded, log)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unpack log")
		}
		task := NewParticipantAddedHandler(*event)
		return []core.Task{task}, nil
	}
	return nil, nil
}
