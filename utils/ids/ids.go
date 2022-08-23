package ids

import (
	"crypto/ecdsa"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	"github.com/avalido/mpc-controller/utils/bytes"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func ShortIDFromPrivKeyHex(privKeyHex string) (*ids.ShortID, error) {
	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return ShortIDFromPubKey(&privKey.PublicKey)
}

func ShortIDFromPubKeyBytes(pubKeyBytes []byte) (*ids.ShortID, error) {
	pubKey, err := myCrypto.UnmarshalPubKeyHex(bytes.BytesToHex(pubKeyBytes))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ShortIDFromPubKey(pubKey)
}

func ShortIDFromPubKeyHex(pubKeyHex string) (*ids.ShortID, error) {
	pubKey, err := myCrypto.UnmarshalPubKeyHex(pubKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ShortIDFromPubKey(pubKey)
}

func ShortIDFromPubKey(pubKey *ecdsa.PublicKey) (*ids.ShortID, error) {
	pubKeyBytes := myCrypto.MarshalPubkey(pubKey)[1:]

	normPubKeyBytes, err := myCrypto.NormalizePubKeyBytes(pubKeyBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	shortID, err := ids.ToShortID(hashing.PubkeyBytesToAddress(normPubKeyBytes))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &shortID, nil
}
