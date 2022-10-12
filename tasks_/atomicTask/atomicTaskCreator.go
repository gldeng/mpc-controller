package atomicTask

import (
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/pool"
	staking "github.com/avalido/mpc-controller/tasks_"
	kbcevents "github.com/kubecost/events"
)

type AtomicTxTaskCreator struct {
	ReqStartedEvtDispatcher kbcevents.Dispatcher[*events.RequestStarted]
	StakeTaskCreator        *staking.StakeTaskCreator

	Pool pool.WorkerPool
}

//func (a *AtomicTxTaskCreator) Init() {
//	a.ReqStartedEvtDispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)
//}

//func (a *AtomicTxTaskCreator) re
//
//
//var reqStartedEvtHandler = func(evt *events.RequestStarted) {
//	stakeReq := evt.JoinedReq.Args.(*storage.StakeRequest)
//	stakeTask, _ := .createStakeTask(stakeReq, *evt.ReqHash)
//	// todo: logic...
//}
//
//var reqStartedEvtFilter = func(evt *events.RequestStarted) bool {
//	return evt.TaskType == storage.TaskTypStake
//}

// decouple
// async, sync?

// todo: subscribe business event so as to create corresponding AtomicTask

func (c *AtomicTxTaskCreator) createAtomicTask() {
	// todo: create AtomicTask to business logic
	var t *AtomicTask

	c.Pool.Submit(t.Do)
}

func (c *AtomicTxTaskCreator) createStakeAtomicTask() *AtomicTask {
	return nil
}

func (c *AtomicTxTaskCreator) createRecoverAtomicTask() *AtomicTask {
	return nil
}
