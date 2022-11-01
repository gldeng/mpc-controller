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
	quorum         types.QuorumInfo
	request        *Request
	signedImportTx *txs.Tx
	taskCtxMock    *mocks.TaskContext
}

func (s *AddDelegatorTestSuite) SetupTest() {
	require := s.Require()

	s.id = "abc"
	s.quorum = types.QuorumInfo{}
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
}

func (s *AddDelegatorTestSuite) TestNext() {
	require := s.Require()
	task, err := NewAddDelegator(s.request, s.id, s.quorum, s.signedImportTx)
	require.Nil(err)
	require.NotNil(task)
}

func TestAddDelegatorTestSuite(t *testing.T) {
	suite.Run(t, new(AddDelegatorTestSuite))
}
