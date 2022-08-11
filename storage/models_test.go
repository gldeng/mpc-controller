package storage

import (
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRequestHash_SetTaskType(t *testing.T) {
	reqHashBytes := bytes.HexTo32Bytes("0xa5b548b8bdfa18ee8cdbc85ac440701634719b87a5a48078bc683e09087508b5")
	reqHash := (RequestHash)(reqHashBytes)
	reqHash.SetTaskType(TaskTypStake)
	require.True(t, reqHash.IsTaskType(TaskTypStake))
	require.Equal(t, TaskType(1), reqHash.TaskType())
}
