package stake

import (
	"context"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/storage"
	kbcevents "github.com/kubecost/events"
)

type TaskCreator struct {
	Ctx    context.Context
	Logger logger.Logger

	PartiPubKey storage.PubKey

	DB storage.DB

	MpcClient core.MpcClient
	TxIssuer  txissuer.TxIssuer

	Network chain.NetworkContext

	Bound transactor.Transactor

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.UTXOFetched]
}

func (c *TaskCreator) Start() error {
	reqStartedEvtHandler := func(evt *events.UTXOFetched) {
		t := Task{
			Ctx:    c.Ctx,
			Logger: c.Logger,

			DB:   c.DB,
			Pool: c.Pool,

			Bound: c.Bound,

			UTXOToRecover: evt,
			PartiPubKey:   c.PartiPubKey,
		}
		c.Pool.Submit(t.Do)
	}

	reqStartedEvtFilter := func(evt *events.UTXOFetched) bool {
		return true
	}

	c.Dispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)

	return nil
}

func (c *TaskCreator) Close() error {
	c.Pool.StopAndWait()
	return nil
}
