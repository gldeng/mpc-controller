package pchainevent

import (
	"github.com/avalido/mpc-controller/core"
	types2 "github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/tasks/p2c"
)

var (
	_ core.PChainEventHandler = &RequestCreator{}
)

type RequestCreator struct{}

func (r RequestCreator) Handle(ctx core.EventHandlerContext, utxoBucket types2.UtxoBucket) ([]core.Task, error) {
	task, err := p2c.NewJoinAndMoveBucket(utxoBucket)
	if err != nil {
		return nil, err
	}

	return []core.Task{task}, nil
}
