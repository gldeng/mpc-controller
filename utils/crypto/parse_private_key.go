package crypto

import (
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
)

func ParsePrivateKeySECP256K1R(key string) (*avaCrypto.PrivateKeySECP256K1R, error) {
	f := avaCrypto.FactorySECP256K1R{}
	signer, err := f.ToPrivateKey(common.Hex2Bytes(key))
	if err != nil {
		return nil, err
	}
	return signer.(*avaCrypto.PrivateKeySECP256K1R), nil
}
