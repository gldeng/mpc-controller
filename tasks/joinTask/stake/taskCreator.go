package stake

import (
	"context"
	"github.com/avalido/mpc-controller/contract/transactor"
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

	Bound transactor.Transactor

	Pool       pool.WorkerPool
	Dispatcher kbcevents.Dispatcher[*events.StakeRequestAdded]
}

func (c *TaskCreator) Start() error {
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
	return nil
}

func (c *TaskCreator) Close() error {
	c.Pool.StopAndWait()
	return nil
}
