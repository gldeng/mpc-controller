package network

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAssetIDAVAX(t *testing.T) {
	id, err := AssetIDAVAX()
	require.Truef(t, id != nil && err == nil, "failed to get AVAX asset id, error: %v", err)
}
