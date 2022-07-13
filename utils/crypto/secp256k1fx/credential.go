package secp256k1fx

import "github.com/ava-labs/avalanchego/vms/secp256k1fx"

func FromSigBytes(sig [65]byte) *secp256k1fx.Credential {
	cred := &secp256k1fx.Credential{
		Sigs: make([][65]byte, 1),
	}
	copy(cred.Sigs[0][:], sig[:])
	return cred
}
