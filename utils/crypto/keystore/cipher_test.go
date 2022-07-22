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
		"156177364ae1ca503767382c1b910463af75371856e90202cb0d706cdce53c33",
		"353fb105bbf9c29cbf46d4c93a69587ac478138b7715f0786d7ae1cc05230878",
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
		Encrypt password:  kbxyZB5TF0x32qWLBeSUqMeTkQ4lO6otcvI6ZBbP0PcwO1t3vR52Cp6I8Pc1C25W
			Encrypt Hex:  83113ac0ec9e9afa9e5a4df6af5a07f3a5f6c6383ba0efecb70cc53c88b6278875839c83cf96af6f7ea7504d69f2196e1d23aa23c8f049b25508eb01fd1f240db97f3bce3f109eac0f012eb17c2009adbab83d5ca4ddf118699fc07ef6c74faf
		Encrypt password:  wIfYQEBMSXaPZ6coe8rYKXfZ1aE9jYj8FylK5W3c3tG8NgSsFCmIvWzk3EJA3Bly
			Encrypt Hex:  5f2f5f616b45a59e06f7ffe47c2b5207559edbdae11a596d5f8c06ed47d11e0399744461f96e7cff9adce8c0bea392d81e570df0da30b977487b24e5408ffcd349059517eb7d6626fc7e3e2483debf062a781dbed7a9a7cff5721c8d1ab8c13f
	*/
}
