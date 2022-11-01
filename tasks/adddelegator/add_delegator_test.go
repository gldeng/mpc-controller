package addDelegator

import (
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

// TODO: implement

type AddDelegatorTestSuite struct {
	suite.Suite
	id             string
	quorum         types.QuorumInfo
	request        *Request
	signedImportTx *txs.Tx
}

func (s *AddDelegatorTestSuite) SetupTest() {
	s.id = "abc"
	s.quorum = types.QuorumInfo{}
	s.signedImportTx = nil
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
