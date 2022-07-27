package work

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/misc"
	"sync"
	"sync/atomic"
	"time"
)

type Workshop struct {
	Id         string
	Logger     logger.Logger
	MaxIdleDur time.Duration // 0 means forever
	TaskChan   chan *Task

	MaxWorkspaces    uint32
	livingWorkspaces uint32

	workspacesMap map[string]*Workspace
	lock          *sync.Mutex
}

func NewWorkshop(logger logger.Logger, maxIdleDur time.Duration, maxWorkspaces uint32, taskChan chan *Task) *Workshop {
	workshop := &Workshop{
		Id:            "workshop_" + misc.NewID(),
		Logger:        logger,
		MaxIdleDur:    maxIdleDur,
		TaskChan:      taskChan,
		MaxWorkspaces: maxWorkspaces,
		workspacesMap: make(map[string]*Workspace),
		lock:          new(sync.Mutex),
	}
	return workshop
}

func (w *Workshop) AddTask(ctx context.Context, t *Task) {
	w.Logger.WarnOnTrue(float64(atomic.LoadUint32(&w.livingWorkspaces)) > float64(w.MaxWorkspaces)*0.8,
		"Living workspaces is close to max allowed number",
		[]logger.Field{{"livingWorkspaces", atomic.LoadUint32(&w.livingWorkspaces)},
			{"maxWorkspaces", w.MaxWorkspaces}}...)

	if atomic.LoadUint32(&w.livingWorkspaces) == 0 {
		w.newWorkspace(ctx)
		w.TaskChan <- t
		return
	}

	if atomic.LoadUint32(&w.livingWorkspaces) == w.MaxWorkspaces {
		w.TaskChan <- t
		return
	}

	var isIdle bool
	w.lock.Lock()
	for _, workspace := range w.workspacesMap {
		if workspace.IsIdle() {
			isIdle = true
			break
		}
	}
	w.lock.Unlock()

	if isIdle {
		w.TaskChan <- t
		return
	}

	w.newWorkspace(ctx)
	w.TaskChan <- t
	return
}

func (w *Workshop) LivingWorkspaces() int {
	return int(atomic.LoadUint32(&w.livingWorkspaces))
}

func (w *Workshop) newWorkspace(ctx context.Context) {
	ws := &Workspace{
		Id:         "workspace_" + misc.NewID(),
		Logger:     w.Logger,
		MaxIdleDur: w.MaxIdleDur,
		TaskChan:   w.TaskChan,
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
