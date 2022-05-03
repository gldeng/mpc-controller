package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
