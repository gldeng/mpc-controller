package p2c

import (
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/pkg/errors"
)

func ExportTxHash(exportTx *txs.ExportTx) ([]byte, error) {

	tx := txs.Tx{
		Unsigned: exportTx,
	}

	unsignedBytes, err := txs.Codec.Marshal(txs.Version, &tx.Unsigned)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func ValidateAndGetCred(hash []byte, sig types.Signature, address ids.ShortID) (*secp256k1fx.Credential, error) {
	sigIndex := 0
	numSigs := 1
	cred := &secp256k1fx.Credential{
		Sigs: make([][SigLength]byte, numSigs),
	}
	copy(cred.Sigs[sigIndex][:], sig[:])

	pk, err := new(avaCrypto.FactorySECP256K1R).RecoverHashPublicKey(hash, sig[:])
	if err != nil {
		return nil, errors.New(ErrMsgPubKeyCannotRecover)
	}
	if address != pk.Address() {
		return nil, errors.New(ErrMsgSignatureInvalid)
	}
	return cred, nil
}

func PackSignedExportTx(unsigned *txs.ExportTx, cred *secp256k1fx.Credential) (*txs.Tx, error) {
	if unsigned == nil {
		return nil, errors.New("missing unsigned tx")
	}
	if cred == nil {
		return nil, errors.New(ErrMsgMissingCredential)
	}

	tx := txs.Tx{
		Unsigned: unsigned,
		Creds: []verify.Verifiable{
			cred,
		},
	}
	err := tx.Sign(txs.Codec, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare signed tx")
	}
	return &tx, nil
}
