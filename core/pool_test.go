package core

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	_ Task        = (*IncrementTask)(nil)
	_ TaskContext = (*MockTaskContext)(nil)
)

type IncrementTask struct {
	Counter int
}

func (i *IncrementTask) GetId() string {
	//TODO implement me
	panic("implement me")
}

func (i *IncrementTask) FailedPermanently() bool {
	//TODO implement me
	panic("implement me")
}

func (i *IncrementTask) IsDone() bool {
	//TODO implement me
	panic("implement me")
}

func (i *IncrementTask) RequiresNonce() bool {
	//TODO implement me
	panic("implement me")
}

func (i *IncrementTask) Next(ctx TaskContext) ([]Task, error) {
	i.Counter = i.Counter + 1
	return nil, nil
}

type MockTaskContext struct {
}

func (m MockTaskContext) GetDb() storage.SlimDb {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) GetEventID(event string) (common.Hash, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) GetMyPublicKey() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) GetParticipantID() storage.ParticipantId {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) GetNetwork() *chain.NetworkContext {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) GetMpcClient() MpcClient {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) IssueCChainTx(txBytes []byte) (ids.ID, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) IssuePChainTx(txBytes []byte) (ids.ID, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) CheckCChainTx(id ids.ID) (TxStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) CheckPChainTx(id ids.ID) (TxStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) NonceAt(account common.Address) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) Emit(event interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m MockTaskContext) GetLogger() logger.Logger {
	//TODO implement me
	panic("implement me")
}

func MockTaskContextFactory() TaskContext {
	return &MockTaskContext{}
}

func TestPool(t *testing.T) {
	p, err := NewExtendedWorkerPool(1, MockTaskContextFactory)
	require.NoError(t, err)
	task := &IncrementTask{Counter: 0}
	p.Submit(task)
	p.Close()
	require.Equal(t, 1, task.Counter)
}
