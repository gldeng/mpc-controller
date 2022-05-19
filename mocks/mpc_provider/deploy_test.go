package mpc_provider

import (
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

// Please make sure a local Avalanche network with url of "http://localhost:9650/ext/bc/C/rpc"
// is available as well as the account holds sufficient AVAX on C-Chain for transaction fee
// before running this testing method to deploy a smart contract.
// To run a local test network please reference "https://docs.avax.network/build/tutorials/platform/create-a-local-test-network/".
// To transfer fund to the address mentioned above, you can use Avalanche wallet from "https://wallet.avax.network/".
// Plus go to "https://docs.avax.network/learn/platform-overview/transaction-fees" for more information on Avalanche transaction fee.
func TestDeployMpcCoordinator(t *testing.T) {
	logger.DevMode = true

	log := logger.Default()
	chainId := big.NewInt(43112)

	privateKey, _ := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")

	cRpcClient := network.DefaultEthClient()

	addr, mpcCoordinator, err := DeployMpcCoordinator(log, chainId, cRpcClient, privateKey)
	require.Nilf(t, err, "error:%v", err)

	spew.Println("Deployed address: ", addr.Hex())
	spew.Dump(mpcCoordinator)
}
