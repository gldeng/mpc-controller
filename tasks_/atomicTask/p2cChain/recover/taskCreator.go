package stake

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	"github.com/dgraph-io/ristretto"
	kbcevents "github.com/kubecost/events"
)

type TaskCreator struct {
	Ctx    context.Context
	Logger logger.Logger

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Network chain.NetworkContext

	ContractCaller caller.Caller

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.RequestStarted]

	UTXOsCache *ristretto.Cache
}

func (c *TaskCreator) Start() {
	reqStartedEvtHandler := func(evt *events.RequestStarted) {
		t := Task{
			Ctx:    c.Ctx,
			Logger: c.Logger,

			Network: c.Network,

			ContractCaller: c.ContractCaller,

			MpcClient: c.MpcClient,
			TxIssuer:  c.TxIssuer,

			Pool:       c.Pool,
			Dispatcher: kbcevents.NewDispatcher[*events.UTXOExported](),

			Joined: evt,

			UTXOsCache: c.UTXOsCache,
		}
		c.Pool.Submit(t.Do)
	}

	reqStartedEvtFilter := func(evt *events.RequestStarted) bool {
		return evt.TaskType == storage.TaskTypRecover
	}

	c.Dispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)
}

func (c *TaskCreator) Close() {
	c.Pool.StopAndWait()
}
