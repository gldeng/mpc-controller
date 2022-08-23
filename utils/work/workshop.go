package work

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/utils/misc"
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Workshop struct {
	Id         string
	Logger     logger.Logger
	MaxIdleDur time.Duration // 0 means forever

	taskChan chan *Task

	TaskZone *TaskZone
	once     sync.Once

	MaxWorkspaces    uint32
	livingWorkspaces uint32

	workspacesMap map[string]*Workspace
	lock          sync.Mutex
}

func NewWorkshop(logger logger.Logger, name string, maxIdleDur time.Duration, maxWorkspaces uint32) *Workshop {
	taskChan := make(chan *Task, 1024)
	taskZone := &TaskZone{
		Id:               name + "_workshop_" + "_taskZone_" + misc.NewID()[:4],
		Logger:           logger,
		PerTaskQueueSize: 1024,
	}

	workshop := &Workshop{
		Id:            name + "_workshop_" + misc.NewID()[:4],
		Logger:        logger,
		MaxIdleDur:    maxIdleDur,
		taskChan:      taskChan,
		TaskZone:      taskZone,
		MaxWorkspaces: maxWorkspaces,
		workspacesMap: make(map[string]*Workspace),
	}
	return workshop
}

func (w *Workshop) AddTask(ctx context.Context, t *Task) error {
	w.once.Do(func() {
		go w.run(ctx)
	})
	return errors.WithStack(w.TaskZone.EnZone(t))
}

func (w *Workshop) run(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if w.TaskZone.IsEmpty() {
				break
			}
			if atomic.LoadUint32(&w.livingWorkspaces) == 0 {
				w.newWorkspace(ctx)
				w.taskChan <- w.TaskZone.DeZone()
				break
			}
			if atomic.LoadUint32(&w.livingWorkspaces) == w.MaxWorkspaces {
				if !w.isIdle() {
					w.Logger.Warn(w.Id + " no idle workspace.")
					break
				}
				w.taskChan <- w.TaskZone.DeZone()
				break
			}
			if w.isIdle() {
				w.taskChan <- w.TaskZone.DeZone()
				break
			}
			w.newWorkspace(ctx)
			w.taskChan <- w.TaskZone.DeZone()
		}
		w.checkState(ctx)
	}
}

func (w *Workshop) checkState(ctx context.Context) {
	w.Logger.WarnOnTrue(float64(atomic.LoadUint32(&w.livingWorkspaces)) > float64(w.MaxWorkspaces)*0.8,
		w.Id+" too many living workspaces",
		[]logger.Field{{"livingWorkspaces", atomic.LoadUint32(&w.livingWorkspaces)},
			{"maxWorkspaces", w.MaxWorkspaces}}...)

	z := w.TaskZone
	var totalTasks int
	var totalCap int
	for priority, q := range z.taskPriorityQueues {
		totalTasks = totalTasks + q.Count()
		totalCap = totalCap + q.Capacity()

		isTooMany := float64(q.Count()) > float64(q.Capacity())*0.8
		z.Logger.WarnOnTrue(isTooMany, z.Id+" too many en-zoned tasks of priority "+strconv.Itoa(priority),
			[]logger.Field{{"totalTasks", q.Count()}, {"maxTasks", q.Capacity()}}...)
	}

	isTooMany := float64(totalTasks) > float64(totalCap)*0.8

	z.Logger.WarnOnTrue(isTooMany, z.Id+" too many en-zoned tasks",
		[]logger.Field{{"totalTasks", totalTasks},
			{"maxTasks", totalCap}}...)

	z.Logger.DebugOnTrue(isTooMany, z.Id+" en-zoned tasks stats in priority", []logger.Field{
		{"p5", z.tasksInQueue(5)},
		{"p4", z.tasksInQueue(4)},
		{"p3", z.tasksInQueue(3)},
		{"p2", z.tasksInQueue(2)},
		{"p1", z.tasksInQueue(1)},
		{"p0", z.tasksInQueue(0)}}...)
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
		Id:         w.Id + "_workspace_" + misc.NewID()[:4],
		Logger:     w.Logger,
		MaxIdleDur: w.MaxIdleDur,
		TaskChan:   w.taskChan,
	}
	w.lock.Lock()
	w.workspacesMap[ws.Id] = ws
	w.lock.Unlock()

	prom.WorkshopWorkspaces.Inc()

	go func() {
		w.Logger.Debug(ws.Id+" workspace opened", []logger.Field{{"openedWorkspace", ws.Id}, {"livingWorkspaces", atomic.LoadUint32(&w.livingWorkspaces) + 1}}...)
		atomic.AddUint32(&w.livingWorkspaces, 1)
		ws.Run(ctx)
	loop:
		old := atomic.LoadUint32(&w.livingWorkspaces)
		if swapped := atomic.CompareAndSwapUint32(&w.livingWorkspaces, old, old-1); swapped {
			w.Logger.Debug(ws.Id+" workspace closed", []logger.Field{{"closedWorkspace", ws.Id}, {"livingWorkspaces", atomic.LoadUint32(&w.livingWorkspaces)}}...)
			w.lock.Lock()
			delete(w.workspacesMap, ws.Id)
			w.lock.Unlock()
			return
		}
		goto loop
		prom.WorkshopWorkspaces.Dec()
	}()
}
