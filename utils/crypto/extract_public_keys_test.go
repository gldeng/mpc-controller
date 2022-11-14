package crypto

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

var Keys = []string{
	"353fb105bbf9c29cbf46d4c93a69587ac478138b7715f0786d7ae1cc05230878",
	"7084300e7059ea4b308ec5b965ef581d3f9c9cd63714082ccf9b9d1fb34d658b",
	"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	"156177364ae1ca503767382c1b910463af75371856e90202cb0d706cdce53c33",
	"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
	"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
	"b17eac91d7aa2bd5fa72916b6c8a35ab06e8f0c325c98067bbc9645b85ce789f",
}

func TestExtractPubKeysForParticipants(t *testing.T) {
	pubKeys, err := ExtractPubKeysForParticipants(Keys)
	require.True(t, pubKeys != nil && err == nil)
	for _, k := range pubKeys {
		fmt.Printf("%x\n", k)
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
