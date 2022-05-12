package config

import (
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseConfig(t *testing.T) {
	logger.DevMode = true

	filename := "./config1.yaml"
	config := ParseConfigFromFile(filename)
	require.NotNil(t, config)

	configStr := `enableDevMode: true
controllerId: "mpc-controller-01"
controllerKey: "59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21"
coordinatorAddress: "0x80D72Ef553C7B5f92f6966c2799d926d26Da3dAF"
mpcServerUrl: "http://localhost:8001"
ethRpcUrl: "http://localhost:9650/ext/bc/C/rpc"
ethWsUrl: "ws://127.0.0.1:9650/ext/bc/C/ws"
cChainIssueUrl: "http://localhost:9650"
pChainIssueUrl: "http://localhost:9650"
confignetwork:
  networkId: 12345
  chainId: 43112
  cChainId: "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU"
  avaxId: "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
  importFee: 1000000
  gasPerByte: 1
  gasPerSig: 1000
  gasFixed: 10000
configdbbadger:
  badgerDbPath: "./mpc_controller_db"
`
	config = ParseConfigFromStr(configStr)
	require.NotNil(t, config)
}

func TestMarshalConfig(t *testing.T) {
	var c ConfigImpl
	bytes, err := yaml.Marshal(c)
	require.Nil(t, err)
	fmt.Println(string(bytes))
}
