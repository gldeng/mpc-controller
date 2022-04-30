package crypto

import (
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SignerTestSuite struct {
	suite.Suite
}

func (suite *SignerTestSuite) SetupTest() {}

func (suite *SignerTestSuite) TestSECP256K1RSigner() {
	require := suite.Require()

	signer, err := NewSECP256K1RSigner()
	require.True(signer != nil && err == nil)

	factory := &avaCrypto.FactorySECP256K1R{}
	privKey, err := factory.NewPrivateKey()
	require.True(privKey != nil && err == nil)

	signer2, err := ToSECP256K1RSigner(privKey.Bytes())
	require.True(signer != nil && err == nil)

	var signers = []Signer{signer, signer2}

	for _, signer := range signers {
		msg := []byte("hello, world")
		sig, err := signer.Sign(msg)
		require.True(sig != nil && err == nil)

		ok := signer.Verify(msg, sig)
		require.True(ok)

		msgHash := hashing.ComputeHash256([]byte("hello, world"))
		sigHash, err := signer.SignHash(msgHash)
		require.True(sigHash != nil && err == nil)

		ok = signer.VerifyHash(msgHash, sigHash)
		require.True(ok)

	}
}

func TestSignerTestSuite(t *testing.T) {
	suite.Run(t, new(SignerTestSuite))
}
