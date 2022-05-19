package crypto

import (
	"crypto/ecdsa"
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
		return nil, errors.Wrapf(err, "Got an error when parse address from public key: %q", pubKeyHex)
	}
	return addr, nil
}

func PubkeysToAddresses(pubkeys []*ecdsa.PublicKey) []*common.Address {
	var addrs = make([]*common.Address, 0)
	for _, pubkey := range pubkeys {
		addrs = append(addrs, PubkeyToAddresse(pubkey))
	}
	return addrs
}

func PubkeyToAddresse(pubkey *ecdsa.PublicKey) *common.Address {
	addr := crypto.PubkeyToAddress(*pubkey)
	return &addr
}

func PubKeyBytesToAddress(b []byte) (*common.Address, error) {
	return PubKeyHexToAddress(common.Bytes2Hex(b))
}

func PubKeyHexToAddress(pubKeyHex string) (*common.Address, error) {
	pubKey, err := UnmarshalPubKeyHex(pubKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	account := crypto.PubkeyToAddress(*pubKey)

	return &account, nil
}
