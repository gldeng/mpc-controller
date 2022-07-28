package work

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/avalido/mpc-controller/utils/misc"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"time"
)

type Workshop struct {
	Id         string
	Logger     logger.Logger
	MaxIdleDur time.Duration // 0 means forever

	taskChan chan *Task
	idleChan chan struct{}

	TaskZone *TaskZone
	once     sync.Once

	MaxWorkspaces    uint32
	livingWorkspaces uint32

	workspacesMap map[string]*Workspace
	lock          sync.Mutex
}

func NewWorkshop(logger logger.Logger, maxIdleDur time.Duration, maxWorkspaces uint32) *Workshop {
	taskChan := make(chan *Task, 256)
	idleChan := make(chan struct{})
	taskZone := &TaskZone{
		Logger:           logger,
		TaskChan:         taskChan,
		IdleChan:         idleChan,
		PerTaskQueueSize: 256,
	}

	workshop := &Workshop{
		Id:            "workshop_" + misc.NewID(),
		Logger:        logger,
		MaxIdleDur:    maxIdleDur,
		taskChan:      taskChan,
		idleChan:      idleChan,
		TaskZone:      taskZone,
		MaxWorkspaces: maxWorkspaces,
		workspacesMap: make(map[string]*Workspace),
	}
	return workshop
}

func (w *Workshop) AddTask(ctx context.Context, t *Task) {
	w.once.Do(func() {
		go w.TaskZone.Run(ctx)
	})

	w.Logger.WarnOnTrue(float64(atomic.LoadUint32(&w.livingWorkspaces)) > float64(w.MaxWorkspaces)*0.8,
		"Too many living workspaces",
		[]logger.Field{{"livingWorkspaces", atomic.LoadUint32(&w.livingWorkspaces)},
			{"maxWorkspaces", w.MaxWorkspaces}}...)

	if atomic.LoadUint32(&w.livingWorkspaces) == 0 {
		w.newWorkspace(ctx)
		w.taskChan <- t
		return
	}

	if atomic.LoadUint32(&w.livingWorkspaces) == w.MaxWorkspaces {
		if !w.isIdle() {
			w.Logger.Warn("No idle workspace, task en-zoned")
			err := backoff.RetryFnExponentialForever(ctx, time.Second, time.Second*10, func() (retry bool, err error) {
				if err := w.TaskZone.EnZone(t); err != nil {
					return true, errors.WithStack(err)
				}
				return false, nil
			})
			w.Logger.ErrorOnError(err, "Failed to en-zone task.")
			return
		}
		w.taskChan <- t
		return
	}

	if w.isIdle() {
		w.taskChan <- t
		return
	}

	w.newWorkspace(ctx)
	w.taskChan <- t
	return
}

func (w *Workshop) LivingWorkspaces() int {
	return int(atomic.LoadUint32(&w.livingWorkspaces))
}

func (w *Workshop) isIdle() bool {
	var isIdle bool
	w.lock.Lock()
	for _, workspace := range w.workspacesMap {
		if workspace.IsIdle() {
			isIdle = true
			break
		}
	}
	w.lock.Unlock()
	return isIdle
}

func (w *Workshop) newWorkspace(ctx context.Context) {
	ws := &Workspace{
		Id:         "workspace_" + misc.NewID(),
		Logger:     w.Logger,
		MaxIdleDur: w.MaxIdleDur,
		TaskChan:   w.taskChan,
		IdleChan:   make(chan struct{}, 32),
	}
	w.lock.Lock()
	w.workspacesMap[ws.Id] = ws
	w.lock.Unlock()

	go func() {
		w.Logger.Debug("Workspace opened", []logger.Field{{"openedWorkspace", ws.Id}, {"livingWorkspaces", atomic.LoadUint32(&w.livingWorkspaces) + 1}}...)
		atomic.AddUint32(&w.livingWorkspaces, 1)
		ws.Run(ctx)
	loop:
		old := atomic.LoadUint32(&w.livingWorkspaces)
		if swapped := atomic.CompareAndSwapUint32(&w.livingWorkspaces, old, old-1); swapped {
			w.Logger.Debug("Workspace closed", []logger.Field{{"closedWorkspace", ws.Id}, {"livingWorkspaces", atomic.LoadUint32(&w.livingWorkspaces)}}...)
			w.lock.Lock()
			delete(w.workspacesMap, ws.Id)
			w.lock.Unlock()
			return
		}
		goto loop
	}()
}
