package types

import (
	"fmt"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/avalido/mpc-controller/utils/crypto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math/big"
	"testing"
)

func TestRequestHash_SetTaskType(t *testing.T) {
	reqHashBytes := bytes.HexTo32Bytes("0xa5b548b8bdfa18ee8cdbc85ac440701634719b87a5a48078bc683e09087508b5")
	reqHash := (RequestHash)(reqHashBytes)
	reqHash.SetTaskType(TaskTypStake)
	require.True(t, reqHash.IsTaskType(TaskTypStake))
	require.Equal(t, TaskType(1), reqHash.TaskType())
}

func TestIndices(t *testing.T) {
	rec := &big.Int{}
	rec.SetBytes(bytes.HexToBytes("0115cfc0993c48371b"))
	indices := Indices(*rec)
	actual := indices.Indices()
	expected := []uint{8, 12, 14, 16, 17, 18, 21, 22, 23, 24, 25, 26, 33, 36, 37, 40, 43, 44, 45, 46, 50, 53, 59, 60, 62, 63, 64}
	require.Equal(t, expected, actual)
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
