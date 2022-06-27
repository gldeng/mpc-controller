package cchain

import (
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/coreth/plugin/evm"
	fx "github.com/avalido/mpc-controller/utils/crypto/secp256k1fx"
	"github.com/pkg/errors"
)

func SignedTx(unsignedTx evm.UnsignedAtomicTx, sig [65]byte) (*evm.Tx, error) {
	tx := &evm.Tx{
		UnsignedAtomicTx: unsignedTx,
		Creds: []verify.Verifiable{
			fx.FromSigBytes(sig),
		},
	}

	err := tx.Sign(evm.Codec, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return tx, nil
}
