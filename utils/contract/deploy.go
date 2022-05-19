package contract

import (
	"context"
	"crypto/ecdsa"
	"github.com/avalido/mpc-controller/logger"
	"github.com/davecgh/go-spew/spew"
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

func Deploy(log logger.Logger, chainId *big.Int, client *ethclient.Client, privKey *ecdsa.PrivateKey, bytecodeJSON string) (*common.Address, error) {
	account := crypto.PubkeyToAddress(privKey.PublicKey)
	nonce, err := client.NonceAt(context.Background(), account, nil)
	if err != nil {
		log.Error("Got an error when query nonce", logger.Field{"error", err})
		return nil, errors.Wrapf(err, "got an error when query nonce")
	}

	bytecodeBytes := common.Hex2Bytes(bytecodeJSON)

	txdata := &types.DynamicFeeTx{
		ChainID:    chainId,
		Nonce:      nonce,
		To:         nil,
		Gas:        uint64(Gas),
		GasFeeCap:  big.NewInt(BaseFee),
		GasTipCap:  big.NewInt(1),
		AccessList: nil,
		Data:       bytecodeBytes,
	}
	tx := types.NewTx(txdata)
	signer := types.LatestSignerForChainID(chainId)
	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privKey)
	if err != nil {
		log.Error("Got an error when sign transaction", logger.Field{"error", err})
		return nil, errors.Wrapf(err, "got an error when sign transaction")
	}

	txSigned, err := tx.WithSignature(signer, signature)
	if err != nil {
		log.Error("Got an error when generate new signed transaction", logger.Field{"error", err})
		return nil, errors.Wrap(err, "got an error when generate new signed transaction")
	}

	err = client.SendTransaction(context.Background(), txSigned)
	if err != nil {
		log.Error("Got an error when send transaction", logger.Field{"error", err})
		return nil, errors.Wrap(err, "got an error when send transaction")
	}
	txHash := txSigned.Hash()

	time.Sleep(5 * time.Second) // NOTE: this sleep value may need to adjust according to different network status.
	rcp, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Error("Got an error when send query transaction receipt", logger.Field{"error", err})
		return nil, errors.Wrapf(err, "Got an error when query transaction receipt.")
	}
	if rcp.Status != 1 {
		log.Error("Failed to deploy contract", logger.Field{"receipt", spew.Sdump(rcp)})
		return nil, errors.Errorf("failed to deploy contract, receipt: %v", spew.Sdump(rcp))
	}
	return &rcp.ContractAddress, nil
}
