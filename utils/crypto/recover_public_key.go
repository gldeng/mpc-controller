package crypto

import (
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	"github.com/ava-labs/coreth/plugin/evm"
)

func RecoverEvmTxPublicKey(tx *evm.Tx) ([]byte, error) {
	err := tx.Sign(evm.Codec, nil)
	if err != nil {
		return nil, err
	}
	unsignedBytes := tx.Bytes()
	hash := hashing.ComputeHash256(unsignedBytes)
	cred := tx.Creds[0].(*secp256k1fx.Credential)
	pk, err := new(avaCrypto.FactorySECP256K1R).RecoverHashPublicKey(hash, cred.Sigs[0][:])
	if err != nil {
		return nil, err
	}
	compressed := pk.Bytes()
	bytes, err := DenormalizePubKeyBytes(compressed)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func RecoverPChainTxPublicKey(tx *txs.Tx) ([]byte, error) {
	err := tx.Sign(txs.Codec, nil)
	if err != nil {
		return nil, err
	}
	unsignedBytes := tx.Unsigned.Bytes()
	hash := hashing.ComputeHash256(unsignedBytes)
	cred := tx.Creds[0].(*secp256k1fx.Credential)
	pk, err := new(avaCrypto.FactorySECP256K1R).RecoverHashPublicKey(hash, cred.Sigs[0][:])
	if err != nil {
		return nil, err
	}
	compressed := pk.Bytes()
	bytes, err := DenormalizePubKeyBytes(compressed)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
