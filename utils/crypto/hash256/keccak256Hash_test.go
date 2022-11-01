package hash256

import (
	"fmt"
	"testing"
)

func TestFromBytes(t *testing.T) {
	hashByts := []byte{152, 182, 204, 78, 53, 30, 219, 145, 13, 242, 93, 19, 210, 187, 88, 218, 133, 153, 11, 80, 78, 164, 122, 114, 254, 223, 243, 179, 57, 82, 57, 94}
	hash := FromBytes(hashByts)
	fmt.Println(hash)
}

func TestFromHex(t *testing.T) {
	hashHex := "0xc02b59f772cb23a75b6ffb9f7602ba25fdd5d8e75ad88efcc013fec2c63b0895"
	hash := FromHex(hashHex)
	fmt.Println(hash)
}
