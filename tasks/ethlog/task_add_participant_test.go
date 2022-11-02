package ethlog

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/testingutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func idFromString(str string) ids.ID {
	id, _ := ids.FromString(str)
	return id
}

func TestAddParticipant(t *testing.T) {

	mpcClient, err := core.NewSimulatingMpcClient("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	config := core.Config{
		Host:              "localhost",
		Port:              9650,
		SslEnabled:        false,
		MpcManagerAddress: common.Address{},
		NetworkContext: chain.NewNetworkContext(
			1337,
			idFromString("2cRHidGTGMgWSMQXVuyqB86onp69HTtw6qHsoHvMjk9QbvnijH"),
			big.NewInt(43112),
			avax.Asset{
				ID: idFromString("BUuypiq2wyuLMvyhzFXcPyxPMCgSp7eeDohhQRqTChoBjKziC"),
			},
			1000000,
			1000000,
			1,
			1000,
			10000,
			300,
		),
		MyPublicKey: common.Hex2Bytes("3217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8bdd1c124d85f6c0b10b6ef24460ff4acd0fc2cd84bd5b9c7534118f472d0c7a1"),
	}

	db := storage.NewInMemoryDb()
	services := core.NewServicePack(config, logger.Default(), mpcClient, db)
	ctx, err := core.NewTaskContextImp(services)
	abi, _ := contract.MpcManagerMetaData.GetAbi()
	//myPubKey := common.Hex2Bytes("3217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8bdd1c124d85f6c0b10b6ef24460ff4acd0fc2cd84bd5b9c7534118f472d0c7a1")
	groupId := common.Hex2Bytes("c9dfdfccdc1a33434ea6494da21cc1e2b03477740c606f0311d1f90665070400")
	var groupId32 [32]byte
	copy(groupId32[:], groupId)
	rawLog := testingutils.MakeEventParticipantAdded(config.MyPublicKey, groupId32, big.NewInt(1))
	event := &contract.MpcManagerParticipantAdded{}
	abi.UnpackIntoInterface(event, "ParticipantAdded", rawLog.Data)
	event.Raw = *rawLog

	handler := NewParticipantAddedHandler(*event)
	next, err := handler.Next(ctx)
	require.Nil(t, next)
	require.NoError(t, err)
	key := []byte("group/")
	key = append(key, groupId...)
	res, err := db.Get(context.Background(), key)
	require.NoError(t, err)
	require.NotNil(t, res)
	group := &types.Group{}
	err = group.Decode(res)
	require.NoError(t, err)
	require.Equal(t, groupId32, group.GroupId)
	require.Equal(t, big.NewInt(1), group.Index)
}
