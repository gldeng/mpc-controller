package keystore

import (
	"fmt"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCipher(t *testing.T) {
	var Keys = []string{
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}
	for _, key := range Keys {
		pss, _ := SecureKey(64)
		fmt.Println("Encrypt password: ", pss)
		encryptBytes := Encrypt(pss, []byte(key))
		fmt.Println("Encrypt Hex: ", bytes.BytesToHex(encryptBytes))
		decryptBytes, err := Decrypt(pss, encryptBytes)
		require.Nil(t, err)
		require.Equal(t, key, string(decryptBytes))
	}

	/*
		Encrypt password:  k0n9MLBofTgo2DRVnxSM9hNw8GD9EZ8YTV3SZXwCNHqAtBzqgPJApCBLk0MvlJHt
		Encrypt Hex:  3db317d8f1ff081a32038c339901f8a6a15f53122dde3b99fa8017a2d0952f5ae5cdf5d0df912b8fc7755776a3b12af8cb7aece27b721d7f2c13b4cc957d1ff40131c7b18472748f2aeff2cf9a8aa46539d93281d367bdbe37179955824810ba
		Encrypt password:  pSCzMBSIKQXt2tOirE71vixMdobWJjhaCVdqm3IXJvwNRZ3r6r8So3IdEhWhPl1U
		Encrypt Hex:  49b5078e82a8ac2da6493fcc8da4ab8d97b3e2ca85e51b0d1fb5d271d4eaa5932a74784ce7672827d7f0ff4aeae1e73cadce9bf0ea75d6dc4eaab7ac8c81cb4a1c034f40d11940ffd65822a9d34f41830303516ac600e42b5e98ccf980efb67e
		Encrypt password:  0gXwUSG7PI4ylyKgL3WnPAF3qWQLnpy0jcu46ha9Fxc74RdylsOli4ZbfJ0e9CPg
		Encrypt Hex:  f4c6214f7a5ec30236b9aaa2cddfc0963a4fafe52b29e2a2f0cf2c246de8fb11546d38591e2e45b9d9076bf6739282dce52b8f93651668875c784a1083adf4ae2af1253372ea921a1e5795975e372e829de2a1753e2d0bafcf760ef984881204
	*/
}
