package crypto

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalizePubKeys(t *testing.T) {
	pubKeyHexs := []string{
		"c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66defafb10f0932dc5c649debab82e6c816e164c7b7ad8abbe974d15a94cd1c2937",
		"d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e5f4d4ffc52db1898fe83742a8732e53322c178acb7113072c8dc6f82bbc00b99",
		"73ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6f06766bde6ff723360cee7f89abab428717f811f460ebf67f5186f75a9f4288d",
	}

	normKeys, err := NormalizePubKeys(pubKeyHexs)
	require.Nil(t, err)
	for _, normKey := range normKeys {
		fmt.Println(normKey)
	}
}
