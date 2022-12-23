package keystore

import (
	"github.com/awnumar/memguard"
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
	lb, err := readFileWithProtection(ks.passwordFile)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}
	defer lb.Destroy()
	err = ks.keystore.Unlock(*ks.account, lb.String())
	return errors.WithStack(err)
}

func (ks *KeyStore) Account() *accounts.Account {
	return ks.account
}

func (ks *KeyStore) EthKeyStore() *ethkeystore.KeyStore {
	return ks.keystore
}

// readFileWithProtection check permissions is '0400' and read file into locked buffer
func readFileWithProtection(file string) (*memguard.LockedBuffer, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve file info")
	}
	if fi.Mode() != 0400 {
		return nil, errors.Errorf("file permission expects 0400, but got %v", fi.Mode())
	}

	fil, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	memguard.CatchInterrupt()
	lb, err := memguard.NewBufferFromReader(fil, int(fi.Size()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create LockedBuffer")
	}

	return lb, nil
}
