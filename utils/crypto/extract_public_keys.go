package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func ExtractPubKeysForParticipants(keys []string) ([][]byte, error) {
	pubKeys := make([][]byte, len(keys))
	for i, k := range keys {
		sk, err := crypto.HexToECDSA(k)
		if err != nil {
			return nil, err
		}
		pubKeys[i] = MarshalPubkey(&sk.PublicKey)[1:]
	}

	return pubKeys, nil
}

func ExtractPubKeysForParticipantsHex(keys []string) ([]string, error) {
	pubKeys := make([]string, len(keys))
	for i, k := range keys {
		sk, err := crypto.HexToECDSA(k)
		if err != nil {
			return nil, err
		}
		pubKeys[i] = common.Bytes2Hex(MarshalPubkey(&sk.PublicKey)[1:])
	}

	return pubKeys, nil
}

func MarshalPubkey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(crypto.S256(), pub.X, pub.Y)
}

func MarshalPubkeys(pubs []*ecdsa.PublicKey) [][]byte {
	var marshaledPubs = make([][]byte, 0)
	for _, pub := range pubs {
		marshaledPubs = append(marshaledPubs, MarshalPubkey(pub))
	}
	return marshaledPubs
}

func PubkeysToAddresses(pubkeys []*ecdsa.PublicKey) []*common.Address {
	var addrs = make([]*common.Address, 0)
	for _, pubkey := range pubkeys {
		addr := crypto.PubkeyToAddress(*pubkey)
		addrs = append(addrs, &addr)
	}
	return addrs
}

func UnmarshalPubKeyHex(pubKeyHex string) (*ecdsa.PublicKey, error) {
	pubKeyBytes := common.Hex2Bytes(pubKeyHex)
	return UnmarshalPubkeyBytes(pubKeyBytes)
}

func UnmarshalPubkeyBytes(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	if pubKeyBytes[0] == 4 {
		x, y := elliptic.Unmarshal(crypto.S256(), pubKeyBytes)
		if x == nil {
			return nil, errors.New("invalid public key")
		}
		return &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}, nil
	} else {
		x, y := secp256k1.DecompressPubkey(pubKeyBytes)
		if x == nil {
			return nil, errors.New("invalid public key")
		}
		return &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}, nil
	}
}
