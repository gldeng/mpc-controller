package ids

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestShortIDFromPrivKeyHex(t *testing.T) {
	privKeyHexArr := []string{
		"56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027",
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}

	for _, privKeyHex := range privKeyHexArr {
		shortID, err := ShortIDFromPrivKeyHex(privKeyHex)
		require.True(t, err == nil && shortID != nil)
		fmt.Println(privKeyHex, ": ", shortID.String())
	}
}

func TestShortIDFromPubKeyHex(t *testing.T) {
	id, err := ShortIDFromPubKeyHex("03384ebafc6f500033058392a5a85438e011b9556486a6687e167a93b307ac1116")
	require.Nil(t, err)
	require.Equal(t, "KyWAshdXvuTDZMsbJntCTS5UYcbJRcGW7", id.String())
}
