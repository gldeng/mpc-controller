package pool

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/stretchr/testify/require"
	"testing"
)

type IncrementTask struct {
	Counter int
}

func (i *IncrementTask) Next(ctx TaskContext) ([]Task, error) {
	i.Counter = i.Counter + 1
	return nil, nil
}

type MockTaskContext struct {
}

func (m MockTaskContext) GetLogger() logger.Logger {
	//TODO implement me
	panic("implement me")
}

func MockTaskContextFactory() TaskContext {
	return &MockTaskContext{}
}

func TestPool(t *testing.T) {
	p, err := New(1, MockTaskContextFactory)
	require.NoError(t, err)
	task := &IncrementTask{Counter: 0}
	p.Submit(task)
	p.Close()
	require.Equal(t, 1, task.Counter)
}
