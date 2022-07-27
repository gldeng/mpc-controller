// todo: add more test cases

package staking

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/tasks/staking/mocks"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"math/big"
	"strconv"
	"testing"
)

type StakeTaskWrapperTestSuite struct {
	suite.Suite
	*StakeTask
}

func (suite *StakeTaskWrapperTestSuite) SetupTest() {
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
		1000000,
		1,
		1000,
		10000,
	)

	taskCreator := &StakeTaskCreator{
		"task-fake-id",
		stakeReq,
		networkCtx,
		"03c237810788fbdce68a497236b5ded1ee817ac7465ccc133a8cfe7b75b35ed1e3",
		0,
	}

	task, err := taskCreator.CreateStakeTask()
	if err != nil {
		panic(errors.Wrap(err, "Failed to create stake task"))
	}
	suite.StakeTask = task
}

func (suite *StakeTaskWrapperTestSuite) TestSignTx() {
	require := suite.Require()

	signRequestArgs := SignRequestArgs{
		TaskID:                 "0x74045c1e4c9c538ddfede3971ba8df6fb9a0de096459ab7f5152334c354ba060",
		CompressedPartiPubKeys: []string{"0373ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6", "03c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66d"},
		CompressedGenPubKeyHex: "03c237810788fbdce68a497236b5ded1ee817ac7465ccc133a8cfe7b75b35ed1e3",
	}

	ctx := context.Background()

	exportTxHash := "3273f531ba059c12f98b4cf7890608c66da392b8d5fc218d6d32041c76fdb674"
	exportTxsignature := "5b12ef4bf066a0d341f1bc4c47f597829a22f7b78dabbe1445a84b13053a2f334d7409466b993916dc0e4285911f21111460ea58da98ebe8fbb752bda74d77f301"
	exportTxSignReq := core.SignRequest{
		SignReqID:              signRequestArgs.TaskID + "-" + strconv.Itoa(0),
		CompressedGenPubKeyHex: signRequestArgs.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: signRequestArgs.CompressedPartiPubKeys,
		Hash:                   exportTxHash,
	}
	mockExportTxSigResultFn := func(ctx context.Context, request *core.SignRequest) *core.Result { // todo: use NewSignDoner() and Expecter Interfaces.
		output := &core.Result{
			MPCReqID:      request.SignReqID,
			Result:        exportTxsignature,
			RequestType:   "SIGN",
			RequestStatus: "DONE",
		}

		return output
	}

	importTxHash := "6c99c91ed157e0c2d7cbd790ed89561b7e4bd71f550d4f176a85122afa90c135"
	importTxsignature := "db5abfea3856cdb6013adae8b8ecd006146158b9d90a31f8aa9ea21b9a08274c71a8f52078236b35a76bc5b62686ae01bfdf585e1785d7be891d6e00d70fc58000"
	importTxSignReq := core.SignRequest{
		SignReqID:              signRequestArgs.TaskID + "-" + strconv.Itoa(1),
		CompressedGenPubKeyHex: signRequestArgs.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: signRequestArgs.CompressedPartiPubKeys,
		Hash:                   importTxHash,
	}
	mockImportTxSigResultFn := func(ctx context.Context, request *core.SignRequest) *core.Result {
		output := &core.Result{
			MPCReqID:      request.SignReqID,
			Result:        importTxsignature,
			RequestType:   "SIGN",
			RequestStatus: "DONE",
		}

		return output
	}

	addDelegatorTxHash := "5884ba92bb58372023c5fa3fe699e152188302f5b1e22ea4b8c9dd623e3b283c"
	addDelegatorTxsignature := "144b6137be716be2e8b22177fc3be4de99c18fc8b9e29fb3eecaa673961fc36e0cbc0449aeca034a10a2f52a31b245b0dcb032d06a317c63fd0adcc3862cb4fe01"
	addDelegatorTxSignReq := core.SignRequest{
		SignReqID:              signRequestArgs.TaskID + "-" + strconv.Itoa(2),
		CompressedGenPubKeyHex: signRequestArgs.CompressedGenPubKeyHex,
		CompressedPartiPubKeys: signRequestArgs.CompressedPartiPubKeys,
		Hash:                   addDelegatorTxHash,
	}
	mockaddDelegatorTxSigResultFn := func(ctx context.Context, request *core.SignRequest) *core.Result {
		output := &core.Result{
			MPCReqID:      request.SignReqID,
			Result:        addDelegatorTxsignature,
			RequestType:   "SIGN",
			RequestStatus: "DONE",
		}

		return output
	}

	signDoner := &mocks.SignDoner{}
	signDoner.On("SignDone", ctx, &exportTxSignReq).Return(mockExportTxSigResultFn, nil)
	signDoner.On("SignDone", ctx, &importTxSignReq).Return(mockImportTxSigResultFn, nil)
	signDoner.On("SignDone", ctx, &addDelegatorTxSignReq).Return(mockaddDelegatorTxSigResultFn, nil)

	signRequester := &SignRequester{
		SignDoner:       signDoner,
		SignRequestArgs: signRequestArgs,
	}

	wrapper := StakeTaskWrapper{
		SignRequester: signRequester,
		StakeTask:     suite.StakeTask,
	}

	err := wrapper.SignTx(ctx)
	require.Nil(err)
}

func TestStakeTaskWrapperTestSuite(t *testing.T) {
	suite.Run(t, new(StakeTaskWrapperTestSuite))
}
