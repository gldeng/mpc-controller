package keystore

import (
	"fmt"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCipher(t *testing.T) {
	key, _ := SecureKey(6)
	fmt.Println("Encrypt Key: ", key)
	dataStr := "59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21"
	encryptBytes := Encrypt(key, []byte(dataStr))
	fmt.Println("Encrypt Hex: ", bytes.BytesToHex(encryptBytes))
	decryptBytes, err := Decrypt(key, encryptBytes)
	require.Nil(t, err)
	require.Equal(t, dataStr, string(decryptBytes))
}
