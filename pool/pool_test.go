package pool

import (
	"github.com/avalido/mpc-controller/chain"
	"github.com/stretchr/testify/require"
	"testing"
)

type IncrementTask struct {
	Counter int
}

func (i *IncrementTask) Next(ctx *TaskContext) ([]Task, error) {
	i.Counter = i.Counter + 1
	return nil, nil
}

func MockTaskContextFactory() *TaskContext {
	return &TaskContext{
		Logger:     nil,
		NonceGiver: nil,
		Network:    chain.NetworkContext{},
		MpcClient:  nil,
		TxIssuer:   nil,
	}
}

func TestPool(t *testing.T) {
	p, err := New(1, MockTaskContextFactory)
	require.NoError(t, err)
	task := &IncrementTask{Counter: 0}
	p.Submit(task)
	p.Close()
	require.Equal(t, 1, task.Counter)
}
