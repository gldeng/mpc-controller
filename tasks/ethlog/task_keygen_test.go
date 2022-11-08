package ethlog

import (
	"context"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/core"
	types2 "github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/utils/testingutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestKeyGeneratedHandler(t *testing.T) {

	mpcClient, err := core.NewSimulatingMpcClient("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	config := core.Config{
		Host:              "localhost",
		Port:              9650,
		SslEnabled:        false,
		MpcManagerAddress: common.Address{},
		NetworkContext: core.NewNetworkContext(
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

	groupId := common.Hex2Bytes("c9dfdfccdc1a33434ea6494da21cc1e2b03477740c606f0311d1f90665070400")
	var groupId32, pId [32]byte
	copy(groupId32[:], groupId)
	copy(pId[:], groupId)
	pId[31] = 1

	db := storage.NewInMemoryDb()
	services := core.NewServicePack(config, logger.Default(), mpcClient, db)
	ctx0, err := core.NewTaskContextImp(services)
	ctx := &TaskContextWrapper{
		participantId: pId,
		inner:         ctx0,
		group: [][]byte{
			common.Hex2Bytes("3217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8bdd1c124d85f6c0b10b6ef24460ff4acd0fc2cd84bd5b9c7534118f472d0c7a1"),
			common.Hex2Bytes("72eab231c150b42e86cbe7398139432d2cad04289a820a922fe17b9d4ba577f4d3c33a90bd5b304344e1bea939ef7d16f428d50d25cada4225d9299d35ef1644"),
			common.Hex2Bytes("73ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6f06766bde6ff723360cee7f89abab428717f811f460ebf67f5186f75a9f4288d"),
			common.Hex2Bytes("8196e06c3e803d0af06693a504ad14317550b4be4396ef57cf5f520c0f84833db8ed1056383ea329b8586cb62c37d80a3d7bb80742bc1bec6d650e6632a62905"),
			common.Hex2Bytes("c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66defafb10f0932dc5c649debab82e6c816e164c7b7ad8abbe974d15a94cd1c2937"),
			common.Hex2Bytes("d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e5f4d4ffc52db1898fe83742a8732e53322c178acb7113072c8dc6f82bbc00b99"),
			common.Hex2Bytes("df7fb5bf5b3f97dffc98ecf8d660f604cad76f804a23e1b6cc76c11b5c92f3456dab26cdf995e6cb7cf772ba892044da9c64b095db7725d9e3c306c484cf54e2"),
		},
	}
	genPubKey := common.Hex2Bytes("c6184cd4d6e7eeadd09410fe06a30bc06355c8c8c4dabd5c1e2d3c30d6ba42386bac735d7f4e7d264ac8741ab382a7868bf1bfa3f3b74a67f83d032309d4599c")

	gKey := []byte("group/")
	gKey = append(gKey, groupId32[:]...)

	group := &types2.Group{
		GroupId:          groupId32,
		Index:            big.NewInt(1),
		MemberPublicKeys: ctx.group,
	}
	gBytes, _ := group.Encode()
	db.Set(context.Background(), gKey, gBytes)

	event := testingutils.MakeEventKeyGenerated(groupId32, genPubKey)

	handler := NewKeyGeneratedHandler(*event)
	next, err := handler.Next(ctx)
	require.Nil(t, next)
	require.NoError(t, err)
	key := []byte("latestPubKey")
	res, err := db.Get(context.Background(), key)
	require.NoError(t, err)
	require.NotNil(t, res)
	pubKey := &types2.MpcPublicKey{}
	err = pubKey.Decode(res)
	require.NoError(t, err)
	require.Equal(t, common.BytesToHash(groupId32[:]), pubKey.GroupId)
	require.Equal(t, genPubKey, []byte(pubKey.GenPubKey))
	require.Equal(t, ctx.group, pubKey.ParticipantPubKeys)
}
