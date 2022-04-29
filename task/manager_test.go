package task

import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/juju/errors"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
	"math/big"
	"testing"
)

const (
	AVAX_ID         = "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
	CCHAIN_ID       = "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU"
	NETWORK_ID      = 12345
	CHAIN_ID        = 43112
	GAS_PER_BYTE    = 1
	GAS_PER_SIG     = 1000
	GAS_FIXED       = 10000
	IMPORT_FEE      = 1000000
	ContractAddress = "0x34c4b2d1b042Eb401c90D7d7b1FC5a6fB33675a5"
)

type TaskManagerTestSuite struct {
	suite.Suite
}

func (suite *TaskManagerTestSuite) SetupTest() {
	logger.DevMode = true
}

func (suite *TaskManagerTestSuite) TestTaskManagerGroup() {
	require := suite.Require()

	type Arg struct {
		privateKey string
		mpc_url    string
	}

	var args = []Arg{
		{privateKey: "59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21", mpc_url: "http://localhost:8001"},
		{privateKey: "6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33", mpc_url: "http://localhost:8002"},
		{privateKey: "5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b", mpc_url: "http://localhost:8003"},
	}

	ctx, shutdown := context.WithCancel(context.Background())
	g, gctx := errgroup.WithContext(ctx)

	// todo: deal with elegant shutdown
	_ = shutdown
	_ = gctx

	for i, arg := range args {
		i := i
		arg := arg
		g.Go(func() error {
			i++
			networkCtx, err := testNetworkContext()
			if err != nil {
				return errors.Annotate(err, "failed to build Avalanche network context")
			}

			sk, err := crypto.HexToECDSA(arg.privateKey)
			if err != nil {
				return errors.Annotatef(err, "failed to parse private key from %q", arg.privateKey)
			}

			mpcClient, err := core.NewMpcClient(arg.mpc_url)
			if err != nil {
				return errors.Annotatef(err, "failed to build mpc-client %d with mpc-url %q", i, arg.mpc_url)
			}

			coordinatorAddr := common.HexToAddress(ContractAddress)
			manager, err := NewTaskManager(i, *networkCtx, mpcClient, sk, coordinatorAddr)
			if err != nil {
				return errors.Annotatef(err, "Failed to build task-manager %d", i)
			}
			logger.Debug("A task manager created", logger.Field{"taskMangerNum", i})

			err = manager.Initialize()
			if err != nil {
				return errors.Annotatef(err, "Failed to initialize task-manager %d", i)
			}

			logger.Debug("Started a task manager",
				logger.Field{"managerID", i},
				logger.Field{"privateKey", arg.privateKey},
				logger.Field{"mpcURL", arg.mpc_url},
				logger.Field{"contractAddress", ContractAddress})
			err = manager.Start()
			if err != nil {
				return errors.Annotatef(err, "Failed to start task-manager %d", i)
			}
			return nil
		})
	}

	err := g.Wait()
	require.Nilf(err, "ERROR STACK: %s", errors.ErrorStack(err))
}

func TestTaskManagerTestSuite(t *testing.T) {
	suite.Run(t, new(TaskManagerTestSuite))
}

func testNetworkContext() (*core.NetworkContext, error) {
	cchainID, err := ids.FromString(CCHAIN_ID)
	if err != nil {
		return nil, err
	}
	assetId, err := ids.FromString(AVAX_ID)
	if err != nil {
		return nil, err
	}

	ctx := core.NewNetworkContext(
		NETWORK_ID,
		cchainID,
		big.NewInt(CHAIN_ID),
		avax.Asset{
			ID: assetId,
		},
		IMPORT_FEE,
		GAS_PER_BYTE,
		GAS_PER_SIG,
		GAS_FIXED,
	)
	return &ctx, nil
}
