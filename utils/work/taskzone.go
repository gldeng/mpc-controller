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
			}
		}
	}
}

func (z *TaskZone) EnZone(t *Task) error {
	_, ok := z.taskPriorityQueues[t.Priority]
	if !ok {
		z.taskPriorityQueues[t.Priority] = queue.NewArrayQueue(z.PerTaskQueueSize)
	}
	q := z.taskPriorityQueues[t.Priority]
	return q.Enqueue(t)
}

func (z *TaskZone) deZone() (t *Task) {
	for priority := 5; priority >= 0; priority-- {
		q := z.taskPriorityQueues[priority]
		if q.Empty() {
			continue
		}
		t = q.Dequeue().(*Task)
		break
	}
	return
}
