package noncer

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoncer(t *testing.T) {
	noncer := New(1, 0)

	nonceGot := noncer.GetNonce(1)
	require.Equal(t, uint64(0), nonceGot)
	nonceGot2 := noncer.GetNonce(1)
	require.Equal(t, nonceGot, nonceGot2)

	nonceGot = noncer.GetNonce(2)
	require.Equal(t, uint64(1), nonceGot)
	nonceGot = noncer.GetNonce(3)
	require.Equal(t, uint64(2), nonceGot)

	ok := noncer.ResetBase(2, 1)
	require.False(t, ok)
	ok = noncer.ResetBase(1, 3)
	require.True(t, ok)

	nonceGot = noncer.GetNonce(5)
	require.Equal(t, uint64(7), nonceGot)
	nonceGot = noncer.GetNonce(2)
	require.Equal(t, uint64(4), nonceGot)
}
