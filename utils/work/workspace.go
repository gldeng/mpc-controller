package work

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Worker interface {
	Do(ctx context.Context, args interface{})
}

type Task struct {
	Args    interface{}
	Ctx     context.Context
	Workers []Worker
}

type Workspace struct {
	Id         string
	MaxIdleDur time.Duration // 0 means forever
	TaskChan   chan *Task

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
			for _, worker := range task.Workers {
				worker := worker
				go func() {
					wg.Add(1)
					worker.Do(task.Ctx, task.Args)
					wg.Done()
				}()
			}
			wg.Wait()
			atomic.StoreUint32(&w.status, 0)
			w.lastActiveTime = time.Now()
		}
	}
}

func (w *Workspace) IsIdle() bool {
	return atomic.LoadUint32(&w.status) == 0
}
