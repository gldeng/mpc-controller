package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFlow(t *testing.T) {
	signers, err := getSigners(testnetKey)
	require.Nil(t, err)

	err = testFlow(signers[0])
	require.Nilf(t, err, "%+v", err)
}
