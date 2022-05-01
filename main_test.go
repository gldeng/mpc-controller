package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFlow(t *testing.T) {
	err := testFlow()
	require.Nilf(t, err, "%+v", err)
}
