package crypto

import (
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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

func (suite *SignerTestSuite) TestVerify() {
	require := suite.Require()

	msg_hash := "6a1be824fa870c1517d9ea84013e75ba81cccb44b48a270f12f1ebe45cb2a0c7"
	pubkey := "029d9175befe99a15fb51929d3181ba9331501330749e8e8e8783ebbd729d385f2"
	sig := "ece2351a119e2778aab76c20c8842eba524e6440f143f4c7a96bc7bb60464449492116ab49298c4e3c1e55586c45fd5917da640504ccafba37dd123bc30cc45b01"

	pubkeyBytes := common.Hex2Bytes(pubkey)
	digestBytes := common.Hex2Bytes(msg_hash)
	sigBytes := common.Hex2Bytes(sig)

	correct := crypto.VerifySignature(pubkeyBytes, digestBytes, sigBytes)

	require.True(correct)
}
