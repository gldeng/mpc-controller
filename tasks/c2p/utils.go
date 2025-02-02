package c2p

import (
	"github.com/ava-labs/avalanchego/ids"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/components/verify"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"math/big"
)

// TODO: Move to correct package
func ToGwei(amount *big.Int) (uint64, error) {
	nAVAXAmount := new(big.Int).Div(amount, big.NewInt(params.GWei))
	if !nAVAXAmount.IsUint64() {
		return 0, errors.New("invalid uint64")
	}
	return nAVAXAmount.Uint64(), nil
}

func ExportTxHash(exportTx *evm.UnsignedExportTx) ([]byte, error) {

	tx := evm.Tx{
		UnsignedAtomicTx: exportTx,
	}
	unsignedBytes, err := evm.Codec.Marshal(txs.Version, &tx.UnsignedAtomicTx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal export tx")
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute hash")
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

func PackSignedExportTx(unsigned *evm.UnsignedExportTx, cred *secp256k1fx.Credential) (*evm.Tx, error) {
	if unsigned == nil {
		return nil, errors.New("missing unsigned tx")
	}
	if cred == nil {
		return nil, errors.New(ErrMsgMissingCredential)
	}

	tx := evm.Tx{
		UnsignedAtomicTx: unsigned,
		Creds: []verify.Verifiable{
			cred,
		},
	}
	err := tx.Sign(evm.Codec, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare signed tx")
	}
	return &tx, nil
}

func ImportTxHash(importTx *txs.ImportTx) ([]byte, error) {
	tx := txs.Tx{
		Unsigned: importTx,
	}
	unsignedBytes, err := platformvm.Codec.Marshal(txs.Version, &tx.Unsigned)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	hash := hashing.ComputeHash256(unsignedBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hash, nil
}

func PackSignedImportTx(unsigned *txs.ImportTx, cred *secp256k1fx.Credential) (*txs.Tx, error) {
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
	err := tx.Sign(platformvm.Codec, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare signed tx")
	}
	return &tx, nil
}
