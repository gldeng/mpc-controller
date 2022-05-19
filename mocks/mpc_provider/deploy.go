package mpc_provider

import (
	"context"
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

func DeployMpcCoordinator(log logger.Logger, chainId *big.Int, client *ethclient.Client, privKey *ecdsa.PrivateKey) (*common.Address, *contract.MpcCoordinator, error) {
	signer, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	log.FatalOnError(err, "Failed to create transaction signer", logger.Field{"error", err})

	addr, tx, mpcCoordinator, err := contract.DeployMpcCoordinator(signer, client)
	log.FatalOnError(err, "Failed to deploy AvaLido smart contract", logger.Field{"error", err})

	time.Sleep(time.Second * 5)
	rcp, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if rcp.Status != 1 {
		log.Fatal("Transaction failed", logger.Field{"receipt", spew.Sdump(rcp)})
		return nil, nil, errors.Errorf("transaction failed, receipt: %s", spew.Sdump(rcp))
	}

	return &addr, mpcCoordinator, nil
}
