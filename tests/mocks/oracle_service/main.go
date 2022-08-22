package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"time"
)

func main() {
	var cChainIdFlag = flag.Int64("cChainId", 43112, "Oracle member private key")
	var cChainUrlFlag = flag.String("cChainUrl", "http://localhost:9650/ext/bc/C/rpc", "C-Chain rpc url")
	var oracleMemberPkFlag = flag.String("oracleMemberPK", "a54a5d692d239287e8358f27caee92ab5756c0276a6db0a062709cd86451a855", "Oracle member private key")
	var oracleManagerAddrFlag = flag.String("oracleManagerAddr", "", "OracleManager contract address")
	flag.Parse()

	logger.DevMode = true
	myLogger := logger.Default()

	contractAddr := common.HexToAddress(*oracleManagerAddrFlag)

	client, err := ethclient.Dial(*cChainUrlFlag)
	if err != nil {
		panic(err)
	}

	oracleManager, err := NewOracleManager(contractAddr, client)
	if err != nil {
		panic(err)
	}

	// Parse private key
	myPrivKey, err := crypto.HexToECDSA(*oracleMemberPkFlag)
	if err != nil {
		panic(err)
	}

	myAddr := addrs.PubkeyToAddresse(&myPrivKey.PublicKey)
	fmt.Printf("Oracle member address: %v\n", myAddr)

	// Create transaction signer
	chainId := big.NewInt(*cChainIdFlag)
	signer, err := bind.NewKeyedTransactorWithChainID(myPrivKey, chainId)
	if err != nil {
		panic(err)
	}

	o := Oracle{myLogger, client, signer, oracleManager}
	for {
		if err := o.ReceiveMemberReport(context.Background()); err != nil {
			myLogger.ErrorOnError(err, "Failed to ReceiveMemberReport")
		}
		myLogger.InfoNilError(err, "Success to call ReceiveMemberReport")
		time.Sleep(time.Hour * 24)
	}
}

type Oracle struct {
	Logger        logger.Logger
	EthClient     *ethclient.Client
	Auth          *bind.TransactOpts
	OracleManager *OracleManager
}

func (o *Oracle) ReceiveMemberReport(ctx context.Context) error {
	err := backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (retry bool, err error) {
		epochId := big.NewInt(123456789)
		tx, err := o.OracleManager.ReceiveMemberReport(o.Auth, epochId, o.validators())
		if err != nil {
			errMsg := err.Error()
			switch {
			case strings.Contains(errMsg, "execution reverted"):
				return false, errors.Wrapf(err, "execution reverted")
			}
			return true, errors.WithStack(err)
		}

		err = backoff.RetryFnExponential10Times(ctx, time.Second, time.Second*10, func() (retry bool, err error) {
			rcpt, err := o.EthClient.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				return true, errors.WithStack(err)
			}

			if rcpt.Status != 1 {
				return true, errors.New("Called ReceiveMemberReport but transaction failed")
			}

			return false, nil
		})
		if err != nil {
			return true, errors.WithStack(err)
		}
		return false, errors.WithStack(err)
	})

	return errors.WithStack(err)
}

func (o *Oracle) validators() []*big.Int {
	var validators []*big.Int
	for i := 0; i < 5; i++ {
		validator := o.packValidator(uint64(i), true, true, 100)
		validators = append(validators, validator)
	}
	return validators
}

func (o *Oracle) packValidator(nodeIndex uint64, hasUptime bool, hasSpace bool, hundredsOfAvax uint64) *big.Int {
	if nodeIndex >= 4096 {
		panic("invalid nodeIndex")
	}
	nodeIndexBytes := new(big.Int).SetUint64(nodeIndex).Bytes()
	nodeIndexBig := new(big.Int).SetBytes(nodeIndexBytes)

	if hundredsOfAvax >= 1024 {
		panic("invalid hundredsOfAvax")
	}
	hundredsOfAvaxBtes := new(big.Int).SetUint64(hundredsOfAvax).Bytes()
	data := new(big.Int).SetBytes(hundredsOfAvaxBtes)

	if hasUptime {
		oneBigLsh23 := new(big.Int).Lsh(big.NewInt(1), 23)
		data = new(big.Int).Or(data, oneBigLsh23)
	}

	if hasSpace {
		oneBigLsh22 := new(big.Int).Lsh(big.NewInt(1), 22)
		data = new(big.Int).Or(data, oneBigLsh22)
	}

	nodeIndexBigLsh10 := new(big.Int).Lsh(nodeIndexBig, 10)
	return new(big.Int).Or(data, nodeIndexBigLsh10)
}
