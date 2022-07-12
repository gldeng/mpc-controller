package pchain

import (
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	fx "github.com/avalido/mpc-controller/utils/crypto/secp256k1fx"
	"github.com/pkg/errors"
)

func Tx(unsignedTx txs.UnsignedTx, sig [65]byte) (*txs.Tx, error) {
	tx := &txs.Tx{
		Unsigned: unsignedTx,
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
