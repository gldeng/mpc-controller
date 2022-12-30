package types

import (
	"fmt"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestRequestHash_SetTaskType(t *testing.T) {
	reqHashBytes := bytes.HexTo32Bytes("0xa5b548b8bdfa18ee8cdbc85ac440701634719b87a5a48078bc683e09087508b5")
	reqHash := (RequestHash)(reqHashBytes)
	reqHash.SetTaskType(TaskTypStake)
	require.True(t, reqHash.IsTaskType(TaskTypStake))
	require.Equal(t, TaskType(1), reqHash.TaskType())
}

// PubKey test suite
type PubKeyTestSuite struct {
	suite.Suite
	pubKey PubKey
}

func (suite *PubKeyTestSuite) SetupTest() {
	compressdGenPubKeyStr := "0241538c55805b2614cce2a9298d4479911409ded5e7ddf042629e76c38c237546" //
	pubKeyBytes, err := crypto.DenormalizePubKeyFromHex(compressdGenPubKeyStr)
	suite.Require().Nil(err)
	suite.pubKey = pubKeyBytes
}

func (suite *PubKeyTestSuite) TestString() {
	pubKeyStr := suite.pubKey.String()
	fmt.Printf("Decomporessed key: %v\n", pubKeyStr)
}

func (suite *PubKeyTestSuite) TestCChainAddress() {
	require := suite.Require()
	cChainAddr, err := suite.pubKey.CChainAddress()
	require.Nil(err)
	fmt.Printf("C-Chain Address: %v\n", cChainAddr.String())
}

func (suite *PubKeyTestSuite) TestPChainAddress() {
	require := suite.Require()
	pChainAddr, err := suite.pubKey.PChainAddress()
	require.Nil(err)
	fmt.Printf("P-Chain Address: %v\n", pChainAddr.String())
}

func TestSignerTestSuite(t *testing.T) {
	suite.Run(t, new(PubKeyTestSuite))
}
