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

type AvaLidoStaker struct {
	log      logger.Logger
	cChainId *big.Int

	cRpcClient *ethclient.Client
	cWsClient  *ethclient.Client

	cRpcAvaLido *contract.AvaLido
	cWsAvaLido  *contract.AvaLido

	cPrivateKey *ecdsa.PrivateKey
	cTxSigner   *bind.TransactOpts
}

func New(log logger.Logger,
	cChainId *big.Int,
	avaLidoAddr *common.Address,
	cPrivateKey *ecdsa.PrivateKey,
	cRpcClient *ethclient.Client,
	cWsClient *ethclient.Client) *AvaLidoStaker {

	rpcAvaLido, err := contract.NewAvaLido(*avaLidoAddr, cRpcClient)
	log.FatalOnError(err, "Failed to create AvaLido bindings", logger.Field{"error", err})
	wsAvaLido, err := contract.NewAvaLido(*avaLidoAddr, cWsClient)
	log.FatalOnError(err, "Failed to create AvaLido bindings", logger.Field{"error", err})

	signer, err := bind.NewKeyedTransactorWithChainID(cPrivateKey, cChainId)
	log.FatalOnError(err, "Failed to create transaction signer", logger.Field{"error", err})

	return &AvaLidoStaker{
		log:         log,
		cChainId:    cChainId,
		cRpcClient:  cRpcClient,
		cWsClient:   cWsClient,
		cRpcAvaLido: rpcAvaLido,
		cWsAvaLido:  wsAvaLido,
		cPrivateKey: cPrivateKey,
		cTxSigner:   signer,
	}
}

func (a *AvaLidoStaker) InitiateStake() error {
	tx, err := a.cRpcAvaLido.InitiateStake(a.cTxSigner)
	if err != nil {
		a.log.Error("Got an error when initiate stake", logger.Field{"error", err})
		return errors.Wrap(err, "got an error when initiate stake")
	}

	time.Sleep(time.Second * 5)
	rcp, err := a.cRpcClient.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		a.log.Error("Got an error when query transaction receipt", logger.Field{"error", err})
		return errors.Wrap(err, "got an error when query transaction receipt")
	}

	if rcp.Status != 1 {
		a.log.Error("Transaction failed", logger.Field{"receipt", spew.Sdump(rcp)})
		return errors.Errorf("transaction failed, receipt: %s", spew.Sdump(rcp))
	}

	return nil
}
