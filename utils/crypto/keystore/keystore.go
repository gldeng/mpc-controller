package keystore

import (
	"github.com/ethereum/go-ethereum/accounts"
	ethkeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"os"
)

type KeyStore struct {
	passwordFile string
	account      *accounts.Account
	keystore     *ethkeystore.KeyStore
}

func New(addr common.Address, passwordFile, keystoreDir string) (*KeyStore, error) {
	var ks KeyStore
	// TODO: check 0400 for safe read
	ks.passwordFile = passwordFile
	ethKeystore := ethkeystore.NewKeyStore(keystoreDir, ethkeystore.StandardScryptN, ethkeystore.StandardScryptP)
	accounts_ := ethKeystore.Accounts()
	for _, account := range accounts_ {
		if account.Address == addr {
			ks.account = &account
			break
		}
	}

	if ks.account == nil {
		return nil, errors.New("found no account in keystore")
	}

	ks.keystore = ethKeystore
	return &ks, nil
}

func (ks *KeyStore) Lock() error {
	err := ks.keystore.Lock(ks.account.Address)
	return errors.WithStack(err)
}

func (ks *KeyStore) Unlock() error {
	// TODO: check 0400 for safe read
	pass, err := os.ReadFile(ks.passwordFile)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}
	err = ks.keystore.Unlock(*ks.account, string(pass))
	return errors.WithStack(err)
}

func (ks *KeyStore) Account() *accounts.Account {
	return ks.account
}

func (ks *KeyStore) EthKeyStore() *ethkeystore.KeyStore {
	return ks.keystore
}
