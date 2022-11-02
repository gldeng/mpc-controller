package addDelegator

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/core/mocks"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/testingutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

type AddDelegatorTestSuite struct {
	suite.Suite
	id         string
	stakeParam *StakeParam
	quorum     types.QuorumInfo

	logger     logger.Logger
	networkCtx *chain.NetworkContext
}

func (s *AddDelegatorTestSuite) SetupTest() {
	require := s.Require()

	// Setup id
	s.id = "0xc02b59f772cb23a75b6ffb9f7602ba25fdd5d8e75ad88efcc013fec2c63b0895"

	// Setup stake param
	nodeID, err := ids.NodeIDFromString("NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5")
	require.Nil(err)
	utxos, err := testingutils.GetRewardUTXOs("http://34.172.25.188:9650", "cxbA4wytAUWTRmNyqfYQHnHdR8vYthyeCrDFWEQULiUHPyVu2")
	require.Nil(err)
	require.NotNil(utxos)
	s.stakeParam = &StakeParam{nodeID, 1663315662, 1694830062, utxos}

	// Setup quorum
	privateKeyStr := "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
	pubkeys, err := crypto.ExtractPubKeysForParticipants([]string{privateKeyStr})
	require.Nil(err)

	compressedPubKey, err := crypto.NormalizePubKeyBytes(pubkeys[0])
	require.Nil(err)

	s.quorum = types.QuorumInfo{
		ParticipantPubKeys: nil,
		PubKey:             compressedPubKey,
	}

	// Setup logger
	logger.DevMode = true
	logger.UseConsoleEncoder = true
	s.logger = logger.Default()

	// Setup network context
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

	s.networkCtx = &networkCtx
}

func (s *AddDelegatorTestSuite) TestNext() {
	// Set task context expectation
	taskCtxMock := mocks.NewTaskContext(s.T())
	taskCtxMock.EXPECT().GetLogger().Return(s.logger)
	taskCtxMock.EXPECT().GetNetwork().Return(s.networkCtx)
	//taskCtxMock.EXPECT().IssuePChainTx() // TODO
	//taskCtxMock.EXPECT().CheckPChainTx() // TODO

	// Create AddDelegator
	require := s.Require()
	task, err := NewAddDelegator(s.id, s.quorum, s.stakeParam)
	require.Nil(err)
	require.NotNil(task)

	// TODO: implement
	taskCtxMock.GetLogger()
	taskCtxMock.GetMpcClient()
	taskCtxMock.GetNetwork()
}

func (s *AddDelegatorTestSuite) TestBuildAndSignTx() {
	require := s.Require()

	// Create necessary mocks
	taskCtxMock := mocks.NewTaskContext(s.T())
	mpcClientMock := mocks.NewMpcClient(s.T())

	// Set mock expectation
	taskCtxMock.EXPECT().GetMpcClient().Return(mpcClientMock)
	taskCtxMock.EXPECT().GetNetwork().Return(s.networkCtx)

	mpcClientMock.EXPECT().Sign(mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("*core.SignRequest")).Return(nil)

	// Create AddDelegator task
	task, _ := NewAddDelegator(s.id, s.quorum, s.stakeParam)

	// Build and sign Tx
	err := task.buildAndSignTx(taskCtxMock)
	require.Nil(err)
	require.NotNil(task.tx)
	require.NotNil(task.signReq)
	// TODO: more check
}

func TestAddDelegatorTestSuite(t *testing.T) {
	suite.Run(t, new(AddDelegatorTestSuite))
}
