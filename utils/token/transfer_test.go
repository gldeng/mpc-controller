package token

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferCChain(t *testing.T) {
	ethClient := network.DefaultEthClient()

	privateKey, err := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	require.Nil(t, err, "failed to parse secp256k1 private key")

	var toKeys = []string{
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	}

	for _, toKey := range toKeys {
		toPrivKey, err := crypto.HexToECDSA(toKey)
		require.Nil(t, err, "failed to parse secp256k1 private key")
		toAddress := crypto.PubkeyToAddress(toPrivKey.PublicKey)
		err = TransferInCChain(ethClient, 43112, privateKey, &toAddress, 100000)
		require.Nil(t, err, "failed to transfer")
		balance, err := ethClient.BalanceAt(context.Background(), toAddress, nil)
		require.Nil(t, err, "failed to query balance")
		fmt.Printf("Balance of %v is %v \n", toAddress, balance)
	}
}
