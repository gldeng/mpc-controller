package mpc_client

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRequestNumber(t *testing.T) {
	reqId := "xxxxxx"

	number := QueryRequestNumber(reqId, TypeKeygen)
	require.Equal(t, 0, number)

	number = QueryRequestNumber(reqId, TypeSign)
	require.Equal(t, 0, number)

	IncreaseRequestNumber(reqId, TypeKeygen)
	number = QueryRequestNumber(reqId, TypeKeygen)
	require.Equal(t, 1, number)

	IncreaseRequestNumber(reqId, TypeSign)
	number = QueryRequestNumber(reqId, TypeSign)
	require.Equal(t, 1, number)
}
