package contract

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

var Gas int64 = 8000000
var BaseFee int64 = 300_000_000_000

func Deploy(chainID int64, client *ethclient.Client, privKey *ecdsa.PrivateKey, bytecodeJSON string) (*common.Address, *common.Hash, error) {

	account := crypto.PubkeyToAddress(privKey.PublicKey)
	nonce, err := client.NonceAt(context.Background(), account, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to query nonce for account %q.", account)
	}

	bytecodeBytes := common.Hex2Bytes(bytecodeJSON)

	chainId := big.NewInt(chainID)
	txdata := &types.DynamicFeeTx{
		ChainID:    chainId,
		Nonce:      nonce,
		To:         nil,
		Gas:        uint64(Gas),
		GasFeeCap:  big.NewInt(BaseFee), // maxgascost = 2.1ether
		GasTipCap:  big.NewInt(1),
		AccessList: nil,
		Data:       bytecodeBytes,
	}
	tx := types.NewTx(txdata)
	signer := types.LatestSignerForChainID(big.NewInt(chainID))
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to sign transaction %v with private key %v.", tx, privKey)
	}

	txSigned, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate new signed transaction")
	}

	err = client.SendTransaction(context.Background(), txSigned)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to send transaction")
	}
	txHash := txSigned.Hash()

	time.Sleep(5 * time.Second)
	rcp, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to query transaction receipt for txHash %q.", txHash.Hex())
	}
	logger.Debug("Deployed a smart contract",
		logger.Field{"txHash", txHash.Hex()},
		logger.Field{"receipt", rcp})

	return &rcp.ContractAddress, &txHash, nil
}
