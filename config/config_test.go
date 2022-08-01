package config

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseFile(t *testing.T) {
	config := ParseFile("./config.yaml")
	require.NotNil(t, config)
	require.Equal(t, true, config.EnableDevMode)
	require.Equal(t, "mpc-controller-01", config.ControllerId)
	require.Equal(t, "e63b74641e6e46660a021e17a350d08c910ab7443272d53298cfdda27fd7cb74d946e05ac193cfaa9a91162c9a0bd5f3754368b0f325e83a56d85dd81fc3960cf3daafbedeb25a7db4e2046caf356a66bafe78d48ec21f5718939c28e2c9f12f", config.ControllerKey)
	require.Equal(t, "0x273487EfaC011cfb62361f7b3E3763A54A03D1d3", config.MpcManagerAddress)
	require.Equal(t, "http://localhost:9000", config.MpcServerUrl)
	require.Equal(t, "http://localhost:9650/ext/bc/C/rpc", config.EthRpcUrl)
	require.Equal(t, "ws://127.0.0.1:9650/ext/bc/C/ws", config.EthWsUrl)
	require.Equal(t, "http://localhost:9650", config.CChainIssueUrl)
	require.Equal(t, "http://localhost:9650", config.PChainIssueUrl)

	require.NotNil(t, config.NetworkConfig)
	require.Equal(t, uint32(12345), config.NetworkConfig.NetworkId)
	require.Equal(t, int64(43112), config.NetworkConfig.ChainId)
	require.Equal(t, "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU", config.NetworkConfig.CChainId)
	require.Equal(t, "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe", config.NetworkConfig.AvaxId)
	require.Equal(t, uint64(1000000), config.NetworkConfig.ImportFee)
	require.Equal(t, uint64(1000000), config.NetworkConfig.ExportFee)
	require.Equal(t, uint64(1), config.NetworkConfig.GasPerByte)
	require.Equal(t, uint64(1000), config.NetworkConfig.GasPerSig)
	require.Equal(t, uint64(10000), config.NetworkConfig.GasFixed)

	require.NotNil(t, config.DatabaseConfig)
	require.Equal(t, "./db/mpc_controller_db1", config.DatabaseConfig.BadgerDbPath)

	require.NotNil(t, config.MonitorConfig)
	require.Equal(t, ":7001", config.MonitorConfig.MetricsServeAddr)

	spew.Dump(config)
}
