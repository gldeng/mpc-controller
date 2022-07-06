package checkPeriod

import (
	"context"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/events"
	"github.com/jinzhu/copier"
	"sync"
	"time"
)

const (
	emitAfterSeconds time.Duration = 5 // todo: maybe make it configurable.
)

// Accept event: *events.StakingTaskDoneEvent

// Emit event: *events.StakingPeriodEndedEvent

type StakingPeriodEndedChecker struct {
	Publisher dispatcher.Publisher

	once            sync.Once
	lock            sync.Mutex
	stakingEventMap map[string]*events.StakingTaskDoneEvent // todo: persistence and restore
}

func (eh *StakingPeriodEndedChecker) Do(ctx context.Context, evtObj *dispatcher.EventObject) {
	switch evt := evtObj.Event.(type) {
	case *events.StakingTaskDoneEvent:
		eh.once.Do(func() {
			eh.stakingEventMap = make(map[string]*events.StakingTaskDoneEvent)
			go func() {
				eh.checkStakingEnded(ctx)
			}()
		})

		eh.lock.Lock()
		eh.stakingEventMap[evt.AddDelegatorTxID.String()] = evt
		eh.lock.Unlock()
	}
}

func (eh *StakingPeriodEndedChecker) checkStakingEnded(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			eh.lock.Lock()
			for txID, evt := range eh.stakingEventMap {
				now := time.Now().Unix()
				if uint64(now) > evt.EndTime {
					newEvt := &events.StakingPeriodEndedEvent{}
					copier.Copy(newEvt, evt)
					time.Sleep(emitAfterSeconds)
					eh.Publisher.Publish(ctx, dispatcher.NewRootEventObject("StakingPeriodEndedChecker", newEvt, ctx))
					delete(eh.stakingEventMap, txID)
				}
			}
			eh.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
