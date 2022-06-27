package pchain

import (
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/pkg/errors"
)

func SignedTx(unsignedTx platformvm.UnsignedTx, sig [65]byte) (*platformvm.Tx, error) {
	cred := &secp256k1fx.Credential{
		Sigs: make([][65]byte, 1),
	}
	copy(cred.Sigs[0][:], sig[:])

	tx := &platformvm.Tx{
		UnsignedTx: unsignedTx,
		Creds: []verify.Verifiable{
			cred,
		},
	}

	err := tx.Sign(platformvm.Codec, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tx, nil
}
