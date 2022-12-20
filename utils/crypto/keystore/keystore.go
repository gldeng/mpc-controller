package keystore

import (
	"github.com/ethereum/go-ethereum/accounts"
	ethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pkg/errors"
)

type KeyStore struct {
	*ethkeystore.KeyStore
	Account    *accounts.Account
	Passphrase string
}

func (ks *KeyStore) Lock() error {
	err := ks.KeyStore.Lock(ks.Account.Address)
	return errors.WithStack(err)
}

func (ks *KeyStore) Unlock() error {
	err := ks.KeyStore.Unlock(*ks.Account, ks.Passphrase)
	return errors.WithStack(err)
}
