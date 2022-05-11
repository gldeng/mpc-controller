package storage

import (
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type BadgerDbTestSuite struct {
	suite.Suite
	path   string
	storer Storer
}

func (suite *BadgerDbTestSuite) SetupTest() {
	logger.DevMode = true
	suite.path = "path.test"
	suite.storer = New(logger.Default(), suite.path)

}

func (suite *BadgerDbTestSuite) TestStore() {
	require := suite.Require()
	defer func() {
		err := suite.storer.Close()
		require.Nil(err)
	}()

	// Store group info
	g := GroupInfo{
		GroupIdHex:     "groupIdHex",
		PartPubKeyHexs: []string{"partPubKeyHexs1", "partPubKeyHexs1"},
		Threshold:      1,
	}

	err := suite.storer.StoreGroupInfo(&g)
	require.Nil(err)

	// Store participant info
	p := &ParticipantInfo{
		PubKeyHashHex: "PubKeyHashHex",
		PubKeyHex:     "PubKeyHex",
		GroupIdHex:    "GroupIdHex",
		Index:         3,
	}
	err = suite.storer.StoreParticipantInfo(p)
	require.Nil(err)

	// Store generated public key info
	gpk := GeneratedPubKeyInfo{
		PubKeyHashHex: "PubKeyHashHex",
		PubKeyHex:     "PubKeyHex",
		GroupIdHex:    "GroupIdHex",
	}
	err = suite.storer.StoreGeneratedPubKeyInfo(&gpk)
	require.Nil(err)

	// Store keygen request info
	kg := KeygenRequestInfo{
		RequestIdHex:     "RequestIdHex",
		GroupIdHex:       "GroupIdHex",
		RequestAddedAt:   time.Now(),
		PubKeyReportedAt: time.Now().Add(time.Second * 5),
		PubKeyHashHex:    "PubKeyHashHex",
	}
	err = suite.storer.StoreKeygenRequestInfo(&kg)
	require.Nil(err)
}

// Note: Run TestLoad after TestStore to pass testing
func (suite *BadgerDbTestSuite) TestLoad() {
	require := suite.Require()
	defer func() {
		err := suite.storer.Close()
		require.Nil(err)
	}()

	// Load group info
	g, err := suite.storer.LoadGroupInfo("groupIdHex")
	require.Nil(err)
	spew.Dump(g)

	// Load participant info
	p, err := suite.storer.LoadParticipantInfo("PubKeyHashHex", "GroupIdHex")
	require.Nil(err)
	spew.Dump(p)

	// Load generated public key info
	gpk, err := suite.storer.LoadGeneratedPubKeyInfo("PubKeyHashHex")
	require.Nil(err)
	spew.Dump(gpk)

	// Load keygen request info
	kg, err := suite.storer.LoadKeygenRequestInfo("RequestIdHex")
	require.Nil(err)
	spew.Dump(kg)
}

// Note: Run TestLoad after TestStore to pass testing
func (suite *BadgerDbTestSuite) TestLoads() {
	require := suite.Require()
	defer func() {
		err := suite.storer.Close()
		require.Nil(err)
	}()

	// Load group infos
	gs, err := suite.storer.LoadGroupInfos()
	require.Nil(err)
	fmt.Println("Results of calling LoadGroupInfos")
	for _, v := range gs {
		spew.Dump(v)
	}

	// Load participant infos
	ps, err := suite.storer.LoadParticipantInfos("PubKeyHashHex")
	require.Nil(err)
	fmt.Println("Results of calling LoadParticipantInfos")
	for _, v := range ps {
		spew.Dump(v)
	}

	// Load generated public key info(s) for given group id(s)
	gpks, err := suite.storer.LoadGeneratedPubKeyInfos([]string{"GroupIdHex", "GroupIdHex2"})
	require.Nil(err)
	fmt.Println("Results of calling LoadGeneratedPubKeyInfos")
	for _, v := range gpks {
		spew.Dump(v)
	}

}

func TestMBadgerDbTestSuite(t *testing.T) {
	suite.Run(t, new(BadgerDbTestSuite))
}
