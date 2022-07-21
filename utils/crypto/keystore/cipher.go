package keystore

import (
	"github.com/pkg/errors"
	"github.com/tidwall/secret"
)

var (
	ErrEncrypt = errors.New("encrypt failed")
	ErrDecrypt = errors.New("decrypt failed")
)

func Encrypt(key string, data []byte) []byte {
	encryptBytes, _ := doEncrypt(key, data)
	return encryptBytes
}

func Decrypt(key string, data []byte) ([]byte, error) {
	return doDecrypt(key, data)
}

func doEncrypt(key string, data []byte) ([]byte, error) {
	bytes, err := secret.Encrypt(key, data)
	if err != nil {
		return nil, errors.WithStack(ErrEncrypt)
	}
	return bytes, nil
}

func doDecrypt(key string, data []byte) ([]byte, error) {
	bytes, err := secret.Decrypt(key, data)
	if err != nil {
		return nil, errors.Wrapf(ErrDecrypt, err.Error())
	}
	return bytes, nil
}
