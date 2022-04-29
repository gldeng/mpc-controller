package crypto

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrivateKeySECP256K1R(t *testing.T) {
	keyStr := "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
	key, err := ParsePrivateKeySECP256K1R(keyStr)
	require.Truef(t, key != nil && err == nil, "failed to parse private key %q, error: %v", key, err)
}
