package atomicTask

import (
	"github.com/avalido/mpc-controller/events"
	staking "github.com/avalido/mpc-controller/tasks_"
	kbcevents "github.com/kubecost/events"
)

type AtomicTxTaskAdapter struct {
	ReqStartedEvtDispatcher kbcevents.Dispatcher[*events.RequestStarted]
	StakeTaskCreator        *staking.StakeTaskCreator
}

//func (a *AtomicTxTaskAdapter) Init() {
//	a.ReqStartedEvtDispatcher.AddFilteredEventHandler(reqStartedEvtHandler, reqStartedEvtFilter)
//}

//func (a *AtomicTxTaskAdapter) re
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
