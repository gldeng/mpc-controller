package token

import (
	"context"
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

// todo: check balance sufficience

func TransferInCChain(client *ethclient.Client, chainId *big.Int, privateKey *ecdsa.PrivateKey, to *common.Address, amount *big.Int) error {
	nonce, err := client.NonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey), nil)
	if err != nil {
		return errors.Wrap(err, "failed to query nonce")
	}

	var gas int64 = 21_000 // todo: this value maybe need to adjust
	var baseFee int64 = 300_000_000_000
	txdata := &types.DynamicFeeTx{
		ChainID:    chainId,
		Nonce:      nonce,
		To:         to,
		Gas:        uint64(gas),
		GasFeeCap:  big.NewInt(baseFee),
		AccessList: nil,
		Value:      amount,
	}
	tx := types.NewTx(txdata)
	signer := types.LatestSignerForChainID(chainId)
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privateKey)
	if err != nil {
		return errors.Wrapf(err, "failed to sign transfer transaction")
	}
	txSigned, err := tx.WithSignature(signer, signature)
	if err != nil {
		return errors.Wrapf(err, "failed to compose signed transfer transaction")
	}

	err = client.SendTransaction(context.Background(), txSigned)
	if err != nil {
		return errors.Wrapf(err, "failed to send transfer transaction")
	}
	logger.Debug("Sent a C-Chain transfer transaction",
		logger.Field{"from", crypto.PubkeyToAddress(privateKey.PublicKey).Hex()},
		logger.Field{"to", to.Hex()},
		logger.Field{"amount", amount.String()},
		logger.Field{"TxHash", txSigned.Hash()})

	time.Sleep(5 * time.Second)
	rcp, err := client.TransactionReceipt(context.Background(), txSigned.Hash())
	if err != nil {
		return errors.Wrapf(err, "failed to query transfer transaction receipt")
	}
	logger.Debug("Queried transaction receipt", logger.Field{"receipt", rcp})
	return nil
}
