package staking

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type StakeTaskCreatorTestSuite struct {
	suite.Suite
}

func (suite *StakeTaskCreatorTestSuite) SetupTest() {}

func (suite *StakeTaskCreatorTestSuite) TestSignTx() {
	require := suite.Require()

	amountBigInt, _ := new(big.Int).SetString("25000000000000000000", 10)

	stakeReq := &contract.MpcManagerStakeRequestStarted{
		RequestId:          big.NewInt(1),
		PublicKey:          hash256.FromHex("0x46b91d758e85596319a847b6513ecd1c9f3cf34581ad7db2de77ebbec3dcac94"),
		ParticipantIndices: []*big.Int{big.NewInt(0), big.NewInt(1)},
		NodeID:             "NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5",
		Amount:             amountBigInt,
		StartTime:          big.NewInt(1655946748),
		EndTime:            big.NewInt(1657156348),
	}

	// Convert C-Chain ID
	cchainID, err := ids.FromString("2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU")
	if err != nil {
		panic(errors.Wrap(err, "Failed to convert C-Chain ID"))
	}

	// Convert chain ID
	chainIdBigInt := big.NewInt(43112)

	// Convert AVAX assetId ID
	assetId, err := ids.FromString("2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe")
	if err != nil {
		panic(errors.Wrap(err, "Failed to convert AVAX assetId"))
	}

	// Create NetworkContext
	networkCtx := chain.NewNetworkContext(
		12345,
		cchainID,
		chainIdBigInt,
		avax.Asset{
			ID: assetId,
		},
		1000000,
		1,
		1000,
		10000,
	)

	taskCreator := &StakeTaskCreator{
		stakeReq,
		networkCtx,
		"03c237810788fbdce68a497236b5ded1ee817ac7465ccc133a8cfe7b75b35ed1e3",
		0,
	}

	task, err := taskCreator.CreateStakeTask()
	require.Nil(err)
	require.NotNil(task)

	require.Equal(stakeReq.NodeID, "NodeID-"+task.NodeID.String())

	nAVAXAmount := new(big.Int).Div(stakeReq.Amount, big.NewInt(1_000_000_000))
	require.Equal(nAVAXAmount.Uint64(), task.DelegateAmt)
}

func TestStakeTaskCreatorTestSuite(t *testing.T) {
	suite.Run(t, new(StakeTaskCreatorTestSuite))
}
