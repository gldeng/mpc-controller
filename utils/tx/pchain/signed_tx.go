package pchain

import (
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	fx "github.com/avalido/mpc-controller/utils/crypto/secp256k1fx"
	"github.com/pkg/errors"
)

func SignedTx(unsignedTx platformvm.UnsignedTx, sig [65]byte) (*platformvm.Tx, error) {
	tx := &platformvm.Tx{
		UnsignedTx: unsignedTx,
		Creds: []verify.Verifiable{
			fx.FromSigBytes(sig),
		},
	}

	err := tx.Sign(platformvm.Codec, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tx, nil
}
