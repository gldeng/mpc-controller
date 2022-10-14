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
	"github.com/dgraph-io/ristretto"
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
	Dispatcher kbcevents.Dispatcher[*events.StakeRequestAdded]

	UTXOsCache *ristretto.Cache
}

func (c *TaskCreator) Start() {
	reqStartedEvtHandler := func(evt *events.StakeRequestAdded) {
		t := Task{
			Ctx:    c.Ctx,
			Logger: c.Logger,

			DB:   c.DB,
			Pool: c.Pool,

			Bound: c.Bound,

			TriggerReq:  evt,
			PartiPubKey: c.PartiPubKey,
		}
		c.Pool.Submit(t.Do)
	}

	reqStartedEvtFilter := func(evt *events.StakeRequestAdded) bool {
		return true
	}

	c.Dispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)
}

func (c *TaskCreator) Close() {
	c.Pool.StopAndWait()
}
