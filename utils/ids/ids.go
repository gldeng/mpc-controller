package ids

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/hashing"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func ShortIDFromPrivKeyHex(privKeyHex string) (*ids.ShortID, error) {
	privKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pubKeyBytes := myCrypto.MarshalPubkey(&privKey.PublicKey)[1:]

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
