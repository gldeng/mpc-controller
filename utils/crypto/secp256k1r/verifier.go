package secp256k1r

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/pkg/errors"
)

type Verifier struct {
	PChainAddress ids.ShortID
	crypto.FactorySECP256K1R
}

func (v *Verifier) VerifySig(hash []byte, sig [65]byte) (bool, error) {
	pk, err := v.RecoverHashPublicKey(hash, sig[:])
	if err != nil {
		return false, errors.Wrapf(err, "failed to recover public key with hash %q and signature %q", bytes.BytesToHex(hash), bytes.Bytes65ToHex(sig))
	}

	if v.PChainAddress != pk.Address() {
		return false, errors.Errorf("expected P-Chain address is %q, but got %q", v.PChainAddress, pk.Address())
	}

	return true, nil
}
