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

func TransferInCChain(client *ethclient.Client, chainId int64, privateKey *ecdsa.PrivateKey, to *common.Address, amount int64) error {
	nonce, err := client.NonceAt(context.Background(), crypto.PubkeyToAddress(privateKey.PublicKey), nil)
	if err != nil {
		return errors.Wrap(err, "failed to query nonce")
	}

	chainId_ := big.NewInt(chainId)
	var gas int64 = 21_000 // todo: this value maybe need to adjust
	var baseFee int64 = 300_000_000_000
	txdata := &types.DynamicFeeTx{
		ChainID:    chainId_,
		Nonce:      nonce,
		To:         to,
		Gas:        uint64(gas),
		GasFeeCap:  big.NewInt(baseFee),
		AccessList: nil,
		Value:      big.NewInt(amount),
	}
	tx := types.NewTx(txdata)
	signer := types.LatestSignerForChainID(big.NewInt(chainId))
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
	logger.Debug("Sent a transfer transaction", logger.Field{"TxHash", txSigned.Hash()})

	time.Sleep(5 * time.Second)
	rcp, err := client.TransactionReceipt(context.Background(), txSigned.Hash())
	if err != nil {
		return errors.Wrapf(err, "failed to query transfer transaction receipt")
	}
	logger.Debug("Queried transaction receipt", logger.Field{"receipt", rcp})
	return nil
}
