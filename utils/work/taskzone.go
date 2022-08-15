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
	PerTaskQueueSize int

	taskPriorityQueues map[int]queue.Queue
	once               sync.Once
	sync.Mutex
}

func (z *TaskZone) EnZone(t *Task) error {
	z.once.Do(func() {
		z.taskPriorityQueues = make(map[int]queue.Queue)
	})
	z.Lock()
	defer z.Unlock()
	_, ok := z.taskPriorityQueues[t.Priority]
	if !ok {
		z.taskPriorityQueues[t.Priority] = queue.NewArrayQueue(z.PerTaskQueueSize)
	}
	q := z.taskPriorityQueues[t.Priority]
	return q.Enqueue(t)
}

func (z *TaskZone) IsEmpty() bool {
	z.Lock()
	defer z.Unlock()
	var total int
	for i := 0; i <= 5; i++ {
		total += z.tasksInQueue(i)
	}
	return total == 0
}

func (z *TaskZone) DeZone() (t *Task) {
	z.Lock()
	defer z.Unlock()
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
