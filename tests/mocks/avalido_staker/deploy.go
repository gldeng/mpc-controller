package avalido_staker

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

// todo: move this responsibility to mpc_provider

func DeployAvaLido(log logger.Logger, chainId *big.Int, client *ethclient.Client, privKey *ecdsa.PrivateKey, MpcManagerAddr *common.Address) (*common.Address, *contract.AvaLido, error) {
	signer, err := bind.NewKeyedTransactorWithChainID(privKey, chainId)
	log.FatalOnError(err, "Failed to create transaction signer")

	addr, tx, avalido, err := contract.DeployAvaLido(signer, client, *MpcManagerAddr)
	log.FatalOnError(err, "Failed to deploy AvaLido smart contract")

	time.Sleep(time.Second * 5)
	rcp, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if rcp.Status != 1 {
		log.Fatal("Transaction failed", logger.Field{"receipt", spew.Sdump(rcp)})
		return nil, nil, errors.Errorf("transaction failed, receipt: %s", spew.Sdump(rcp))
	}

	return &addr, avalido, nil
}
