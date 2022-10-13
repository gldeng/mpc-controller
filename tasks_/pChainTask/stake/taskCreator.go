package stake

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/utils/noncer"
	kbcevents "github.com/kubecost/events"
)

type TaskCreator struct {
	Ctx    context.Context
	Logger logger.Logger

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	NonceGiver noncer.Noncer
	Network    chain.NetworkContext

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeAtomicTaskDone]
}

func (c *TaskCreator) Start() {
	reqStartedEvtHandler := func(evt *events.StakeAtomicTaskDone) {
		t := Task{
			Ctx:    c.Ctx,
			Logger: c.Logger,

			Network: c.Network,

			MpcClient: c.MpcClient,
			TxIssuer:  c.TxIssuer,

			Pool:       c.Pool,
			Dispatcher: kbcevents.NewDispatcher[*events.StakeAddDelegatorTaskDone](),

			Atomic: evt,
		}
		c.Pool.Submit(t.Do)
	}

	reqStartedEvtFilter := func(evt *events.StakeAtomicTaskDone) bool {
		return evt.UTXOsToStake != nil
	}

	c.Dispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)
}

func (c *TaskCreator) Close() {
	c.Pool.StopAndWait()
}
