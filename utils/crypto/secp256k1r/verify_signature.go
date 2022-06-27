package secp256k1r

import (
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/pkg/errors"
)

func VerifySignature(addr ids.ShortID, hash []byte, sig [65]byte) (bool, error) {
	factory := avaCrypto.FactorySECP256K1R{}
	pk, err := factory.RecoverHashPublicKey(hash, sig[:])
	if err != nil {
		return false, errors.WithStack(err)
	}
	if addr != pk.Address() {
		return false, errors.New("invalid signature")
	}

	return true, nil
}
