package secp256k1r

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/pkg/errors"
)

type Verifier struct {
	PChainAddress ids.ShortID
	crypto.FactorySECP256K1R
}

func (v *Verifier) VerifySig(hash []byte, sig [65]byte) (bool, error) {
	pk, err := v.RecoverHashPublicKey(hash, sig[:])
	if err != nil {
		return false, errors.WithStack(err)
	}

	if v.PChainAddress != pk.Address() {
		return false, errors.New("invalid signature")
	}

	return true, nil
}
