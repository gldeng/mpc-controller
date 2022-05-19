package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

// Normalize public key(s)

func NormalizePubKeyBytes(pubKeyBytes []byte) ([]byte, error) {
	pubkeyHex, err := NormalizePubKey(common.Bytes2Hex(pubKeyBytes))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return common.Hex2Bytes(*pubkeyHex), nil
}

func NormalizePubKeys(pubKeyHexs []string) ([]string, error) {
	var out []string
	for _, hex := range pubKeyHexs {
		normalized, err := NormalizePubKey(hex)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to normalized public key %q", hex)
		}
		out = append(out, *normalized)
	}
	return out, nil
}

func NormalizePubKey(pubKeyHex string) (*string, error) {
	pubKeyBytes := common.Hex2Bytes(pubKeyHex)
	pubKeyHex0 := pubKeyHex[0]

	switch {
	case len(pubKeyBytes) == 33 && (pubKeyHex0 == 3) || (pubKeyHex0 == 2):
		return &pubKeyHex, nil
	case len(pubKeyBytes) == 65 && pubKeyHex0 == 4:
		compressed, err := toCompressed(pubKeyBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to compress public key %q", pubKeyHex)
		}
		pubN := common.Bytes2Hex(compressed)
		return &pubN, nil
	case len(pubKeyBytes) == 64:
		var newPub [65]byte
		newPub[0] = 4
		copy(newPub[1:], pubKeyBytes)
		compressed, err := toCompressed(newPub[:])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to compress public key %q", pubKeyHex)
		}
		pubN := common.Bytes2Hex(compressed)
		return &pubN, nil
	}

	return nil, errors.Errorf("%q is invalid secp256k1 public key hex", pubKeyHex)
}

func toCompressed(pubKeyBytes []byte) ([]byte, error) {
	x, y := elliptic.Unmarshal(crypto.S256(), pubKeyBytes)
	if x == nil {
		return nil, errors.New("failed to unmarshal public key bytes")
	}
	pk := &ecdsa.PublicKey{Curve: crypto.S256(), X: x, Y: y}
	pubCompressed := elliptic.MarshalCompressed(crypto.S256(), pk.X, pk.Y)
	return pubCompressed, nil
}

// Denormalize public key(s)

func DenormalizePubKeysFromHexs(pubKeys []string) ([][]byte, error) {
	var results [][]byte
	for _, pubKey := range pubKeys {
		result, err := DenormalizePubKeyFromHex(pubKey)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		results = append(results, result)
	}
	return results, nil
}

func DenormalizePubKeyFromHex(pubKeyHex string) ([]byte, error) {
	pubKey, err := UnmarshalPubKeyHex(pubKeyHex)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return DenormalizePubKey(pubKey), nil
}

func DenormalizePubKeys(pubKeys []*ecdsa.PublicKey) [][]byte {
	var results [][]byte
	for _, pubKey := range pubKeys {
		results = append(results, DenormalizePubKey(pubKey))
	}
	return results
}

func DenormalizePubKey(pubKey *ecdsa.PublicKey) []byte {
	pubBytes := crypto.FromECDSAPub(pubKey)
	return pubBytes[1:]
}
