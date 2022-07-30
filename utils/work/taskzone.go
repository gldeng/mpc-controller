package work

import (
	"context"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/queue"
	"sync"
)

type Task struct {
	Args     interface{}
	Ctx      context.Context
	WorkFns  []WorkFn
	Priority int // from low to high: 0->1->2->3->4->5
}

type TaskZone struct {
	Id               string
	Logger           logger.Logger
	TaskChan         chan *Task
	IdleChan         chan struct{}
	PerTaskQueueSize int

	taskPriorityQueues map[int]queue.Queue
	once               sync.Once
}

func (z *TaskZone) Run(ctx context.Context) {
	z.once.Do(func() {
		z.taskPriorityQueues = make(map[int]queue.Queue)
	})
	for {
		select {
		case <-ctx.Done():
			return
		case <-z.IdleChan:
			if t := z.deZone(); t != nil {
				z.TaskChan <- t
				z.Logger.Debug(z.Id + "task de-zoned")
			}
		}
	}
}

func (z *TaskZone) EnZone(t *Task) error {
	defer func() {
		z.Logger.Debug(z.Id+" en-zoned tasks stats in priority", []logger.Field{
			{"p5", z.tasksInQueue(5)},
			{"p4", z.tasksInQueue(4)},
			{"p3", z.tasksInQueue(3)},
			{"p2", z.tasksInQueue(2)},
			{"p1", z.tasksInQueue(1)},
			{"p0", z.tasksInQueue(0)}}...)
	}()

	_, ok := z.taskPriorityQueues[t.Priority]
	if !ok {
		z.taskPriorityQueues[t.Priority] = queue.NewArrayQueue(z.PerTaskQueueSize)
	}
	q := z.taskPriorityQueues[t.Priority]
	return q.Enqueue(t)
}

func (z *TaskZone) deZone() (t *Task) {
	for priority := 5; priority >= 0; priority-- {
		q, ok := z.taskPriorityQueues[priority]
		if !ok {
			continue
		}
		if q.Empty() {
			continue
		}
		t = q.Dequeue().(*Task)
		break
	}
	return
}

func (z *TaskZone) tasksInQueue(priority int) int {
	q, ok := z.taskPriorityQueues[priority]
	if !ok {
		return 0
	}
	return q.Count()
}
