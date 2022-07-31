package work

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"sync"
	"sync/atomic"
	"time"
)

type WorkFn func(ctx context.Context, args interface{})

type Workspace struct {
	Id         string
	Logger     logger.Logger
	MaxIdleDur time.Duration // 0 means forever
	TaskChan   chan *Task
	IdleChan   chan struct{}

	lastActiveTime time.Time
	status         uint32 // 0: idle, 1: busy
}

func (w *Workspace) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.MaxIdleDur == 0 {
				break
			}
			if w.status == 0 {
				if w.lastActiveTime.Add(w.MaxIdleDur).Before(time.Now()) {
					return
				}
			}
		case task := <-w.TaskChan:
			atomic.StoreUint32(&w.status, 1)
			wg := new(sync.WaitGroup)
			for _, workFn := range task.WorkFns {
				wg.Add(1)
				workFn := workFn
				go func() {
					workFn(task.Ctx, task.Args)
					wg.Done()
				}()
			}
			wg.Wait()
			atomic.StoreUint32(&w.status, 0)
			w.lastActiveTime = time.Now()
			w.IdleChan <- struct{}{}
		}
	}
}

func (w *Workspace) IsIdle() bool {
	return atomic.LoadUint32(&w.status) == 0
}
