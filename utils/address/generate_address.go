package address

import (
	"crypto/ecdsa"
	crypto2 "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

// todo: this function does not work right all the time

func EthPubkeyHexToAddress(pubKeyHex string) (*common.Address, error) {
	bytes := common.Hex2Bytes(pubKeyHex)
	compressedBytes := make([]byte, 33)

	compressedBytes[0] = 2
	copy(compressedBytes[1:], bytes[:32])

	addr, err := PubKeyBytesToAddress(compressedBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "parse public key: %q", pubKeyHex)
	}
	return addr, nil
}

func PubkeyToAddresse(pubkey *ecdsa.PublicKey) *common.Address {
	addr := crypto.PubkeyToAddress(*pubkey)
	return &addr
}

func PubKeyBytesToAddress(b []byte) (*common.Address, error) {
	return PubKeyHexToAddress(common.Bytes2Hex(b))
}

func PubKeyHexToAddress(pubKeyHex string) (*common.Address, error) {
	pubKey, err := crypto2.UnmarshalPubKeyHex(pubKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	account := crypto.PubkeyToAddress(*pubKey)

	return &account, nil
}
