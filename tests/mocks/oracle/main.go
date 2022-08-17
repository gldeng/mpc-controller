package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"math/big"
	"strings"
	"time"
)

func main() {
	var cChainIdFlag = flag.Int64("cChainIdFlag", 43112, "Oracle member private key")
	var cChainUrlFlag = flag.String("cChainRpcUrl", "http://localhost:9650/ext/bc/C/rpc", "C-Chain rpc url")
	var oracleMemberPkFlag = flag.String("oracleMemberPkFlag", "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027", "Oracle member private key")
	var oracleManagerAddrFlag = flag.String("oracleManagerAddr", "", "OracleManager contract address")
	flag.Parse()

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

	o := Oracle{client, signer, oracleManager}
	for {
		if err := o.ReceiveMemberReport(context.Background()); err != nil {
			log.Error("Failed to ReceiveMemberReport:%+v\n", err)
		}
		time.Sleep(time.Hour * 24)
	}
}

type Oracle struct {
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
				return true, errors.New("Transaction sent but failed")
			}

			return false, nil
		})
		return false, errors.WithStack(err)
	})

	return errors.WithStack(err)
}

func (o *Oracle) validators() []*big.Int {
	validator := o.packValidator(0, true, true, 100)
	return []*big.Int{validator}
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
