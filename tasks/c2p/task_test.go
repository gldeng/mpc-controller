package c2p

import (
	"encoding/json"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

type MockTaskContext struct {
	net *chain.NetworkContext
}

func (m *MockTaskContext) IssueCChainTx(txBytes []byte) (ids.ID, error) {
	return [32]byte{}, nil
}

func (m *MockTaskContext) IssuePChainTx(txBytes []byte) (ids.ID, error) {
	return [32]byte{}, nil
}

func (m *MockTaskContext) CheckCChainTx(id ids.ID) (pool.Status, error) {
	return pool.Committed, nil
}

func (m *MockTaskContext) CheckPChainTx(id ids.ID) (pool.Status, error) {
	return pool.Committed, nil
}

func NewMockTaskContext() pool.TaskContext {
	net := chain.NewNetworkContext(
		123,
		ids.GenerateTestID(),
		big.NewInt(1),
		avax.Asset{
			ID: ids.GenerateTestID(),
		},
		100,
		100,
		1,
		20,
		100)
	return &MockTaskContext{net: &net}
}

func (m *MockTaskContext) GetLogger() logger.Logger {
	//TODO implement me
	panic("implement me")
}

func (m *MockTaskContext) GetNetwork() *chain.NetworkContext {
	return m.net
}

func (m *MockTaskContext) GetMpcClient() core.MpcClient {
	//TODO implement me
	panic("implement me")
}

func (m *MockTaskContext) NonceAt(account common.Address) (uint64, error) {
	return 1, nil
}

func (m *MockTaskContext) Emit(event interface{}) {
	//TODO implement me
	panic("implement me")
}

func TestTransferC2P(t *testing.T) {
	task, err := New("test", *big.NewInt(100), QuorumInfo{
		ParticipantPubKeys: [][]byte{
			[]byte("123"),
		},
		PubKey: []byte("abc"),
	})
	require.NoError(t, err)
	ctx := NewMockTaskContext()
	task.Next(ctx)
	jstr, _ := json.Marshal(task)
	fmt.Printf(string(jstr))
}
