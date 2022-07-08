package crypto

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

var Keys = []string{
	"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
	"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
	"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
}

func TestExtractPubKeysForParticipants(t *testing.T) {
	pubKeys, err := ExtractPubKeysForParticipants(Keys)
	require.True(t, pubKeys != nil && err == nil)
	for _, k := range pubKeys {
		fmt.Println(k)
	}
}

func TestExtractPubKeysForParticipantsHex(t *testing.T) {
	pubKeys, err := ExtractPubKeysForParticipantsHex(Keys)
	require.True(t, pubKeys != nil && err == nil)
	for _, k := range pubKeys {
		fmt.Println(k)
	}
}

func TestUnmarshalPubKeyHex(t *testing.T) {
	pubKeys, err := ExtractPubKeysForParticipantsHex(Keys)
	require.True(t, pubKeys != nil && err == nil)
	for _, k := range pubKeys {
		fmt.Println(k)
	}
}
