package stake

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
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
	Dispatcher kbcevents.Dispatcher[*events.RequestStarted]
}

func (c *TaskCreator) Init() {
	reqStartedEvtHandler := func(evt *events.RequestStarted) {
		t := Task{
			Ctx:    c.Ctx,
			Logger: c.Logger,

			NonceGiver: c.NonceGiver,
			Network:    c.Network,

			MpcClient: c.MpcClient,
			TxIssuer:  c.TxIssuer,

			Pool:       c.Pool,
			Dispatcher: kbcevents.NewDispatcher[*events.StakeAtomicTaskDone](),

			Joined: evt,
		}
		c.Pool.Submit(t.Do)
	}

	reqStartedEvtFilter := func(evt *events.RequestStarted) bool {
		return evt.TaskType == storage.TaskTypStake
	}

	c.Dispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)
}

func (c *TaskCreator) Close() {
	c.Pool.StopAndWait()
}
