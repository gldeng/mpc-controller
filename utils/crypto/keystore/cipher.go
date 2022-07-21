package keystore

import "github.com/pkg/errors"

var (
	ErrDecrypt = errors.New("decrypt failed")
)

func Encrypt(plaintext string) (pss, ciphertext string) {
	pss = randomPass()
	ciphertext = doEncrypt(pss, plaintext)
	return
}

func Decrypt(pss, ciphertext string) (plaintext string, err error) {
	return doDecrypt(pss, ciphertext)
}

// todo: enhance security of password generation
func randomPass() string {
	return "QrfV2_PsW"
}

// todo: concrete implementation with safe crypto algorithm
func doEncrypt(pass, plaintext string) (ciphertext string) {
	return plaintext
}

// todo: concrete implementation with safe crypto algorithm
func doDecrypt(pss, ciphertext string) (plaintext string, err error) {
	return ciphertext, nil
}
