package addDelegator

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/core/mocks"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

// TODO: implement

type AddDelegatorTestSuite struct {
	suite.Suite
	id             string
	stakeParam     *StakeParam
	signedImportTx *txs.Tx // TODO:
	taskCtxMock    *mocks.TaskContext
	quorum         types.QuorumInfo
}

func (s *AddDelegatorTestSuite) SetupTest() {
	require := s.Require()

	s.id = "abc"
	s.signedImportTx = nil
	taskCtxMock := mocks.NewTaskContext(s.T())

	logger.DevMode = true
	logger.UseConsoleEncoder = true
	taskCtxMock.EXPECT().GetLogger().Return(logger.Default())

	mpcClient, err := core.NewSimulatingMpcClient("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	require.Nil(err)
	taskCtxMock.EXPECT().GetMpcClient().Return(mpcClient)

	cchainID, err := ids.FromString("2cRHidGTGMgWSMQXVuyqB86onp69HTtw6qHsoHvMjk9QbvnijH")
	require.Nil(err)

	assetId, err := ids.FromString("BUuypiq2wyuLMvyhzFXcPyxPMCgSp7eeDohhQRqTChoBjKziC")
	require.Nil(err)

	networkCtx := chain.NewNetworkContext(
		1337,
		cchainID,
		big.NewInt(43112),
		avax.Asset{
			ID: assetId,
		},
		1000000,
		1000000,
		1,
		1000,
		10000,
		300,
	)
	taskCtxMock.EXPECT().GetNetwork().Return(&networkCtx)
	s.taskCtxMock = taskCtxMock

	s.quorum = types.QuorumInfo{
		ParticipantPubKeys: nil,
		PubKey:             mpcClient.UncompressedPublicKeyBytes(),
	}

	nodeID, err := ids.ShortFromString("NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5")
	require.Nil(err)
	s.stakeParam = &StakeParam{ids.NodeID(nodeID), 1663315662, 1694830062, nil}
}

func (s *AddDelegatorTestSuite) TestNext() {
	require := s.Require()
	task, err := NewAddDelegator(s.id, s.quorum, s.stakeParam)
	require.Nil(err)
	require.NotNil(task)
}

func TestAddDelegatorTestSuite(t *testing.T) {
	suite.Run(t, new(AddDelegatorTestSuite))
}
