package work

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"sync"
	"sync/atomic"
	"time"
)

type WorkFn func(ctx context.Context, args interface{})

type Task struct {
	Args    interface{}
	Ctx     context.Context
	WorkFns []WorkFn
}

type Workspace struct {
	Id         string
	Logger     logger.Logger
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
			taskCp := task
			err := backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (retry bool, err error) {
				if atomic.LoadUint32(&w.status) == 1 {
					w.Logger.Debug("Workspace is busy", []logger.Field{{"workspace", w.Id}}...)
					return true, nil
				}
				return false, nil
			})

			if err != nil {
				w.Logger.ErrorOnError(err, "Failed to wait for busy workspace for new task", []logger.Field{{"workspace", w.Id}}...)
				break
			}

			go func() {
				atomic.StoreUint32(&w.status, 1)
				wg := new(sync.WaitGroup)
				for _, workFn := range taskCp.WorkFns {
					workFn := workFn
					go func() {
						wg.Add(1)
						workFn(taskCp.Ctx, taskCp.Args)
						wg.Done()
					}()
				}
				wg.Wait()
				atomic.StoreUint32(&w.status, 0)
				w.lastActiveTime = time.Now()
			}()
		}
	}
}

func (w *Workspace) IsIdle() bool {
	return atomic.LoadUint32(&w.status) == 0
}
