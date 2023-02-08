package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/utils/backoff"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/urfave/cli/v2"
	"math/big"
	"os"
	"time"
)

var (
	fnHost                 = "host"
	fnPort                 = "port"
	fnAvalidoAddress       = "avalidoAddress"
	fnOracleManagerAddress = "oracleManagerAddress"
	fnOracleAddress        = "oracleAddress"
)

var (
	defaultLogger logger.Logger
	tenEthers     *big.Int
)

func init() {
	defaultLogger = logger.Default()
	tenEthers = big.NewInt(10)
	tenEthers.Mul(tenEthers, big.NewInt(params.Ether))
}

func main() {

	logger.DevMode = true
	logger.UseConsoleEncoder = true

	app := &cli.App{
		Name:  "setup",
		Usage: "Handles the MPC operations needed for Avalanche",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        fnHost,
				Required:    true,
				Usage:       "The host of the avalanche rpc service.",
				DefaultText: "localhost",
			},
			&cli.IntFlag{
				Name:        fnPort,
				Required:    true,
				Usage:       "The port of the avalanche rpc service.",
				DefaultText: "9650",
			},
			&cli.StringFlag{
				Name:     fnAvalidoAddress,
				Required: true,
				Usage:    "The address of the deployed AvaLido contract.",
			},
			&cli.StringFlag{
				Name:     fnOracleManagerAddress,
				Required: true,
				Usage:    "The oracle manager address.",
			},
			&cli.StringFlag{
				Name:     fnOracleAddress,
				Required: true,
				Usage:    "The oracle address.",
			},
		},
		Action: runSetup,
	}

	fmt.Printf("Starting process: %v\n", os.Getpid())

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Failed to run controller, error: %+v", err)
		os.Exit(1)
	}
}

func runSetup(c *cli.Context) error {
	config := NewConfig(c.String(fnHost), int16(c.Int(fnPort)),
		common.HexToAddress(c.String(fnAvalidoAddress)), common.HexToAddress(c.String(fnOracleManagerAddress)),
		common.HexToAddress(c.String(fnOracleAddress)))
	signer, err := bind.NewKeyedTransactorWithChainID(oracleAndMpcAdmin, config.GetChainId())
	panicIfError(err)
	config.GetMpcManager().RequestKeygen(signer, mpcGroupId)
	step1FundAccounts(&config)
	step2InitOracle(&config)
	step3SetupOracle(&config)
	step4SetupMpc(&config)
	return nil
}

func step1FundAccounts(config *Config) {
	doFund(config, oracleAndMpcAdmin)
	for _, operator := range mpcOperators {
		doFund(config, operator)
	}
	for _, operator := range oracleOperators {
		doFund(config, operator)
	}
}

func step2InitOracle(config *Config) {
	doSetOracleAddress(config, oracleAndMpcAdmin, config.OracleAddress)
	doSetEpochDuration(config, oracleAndMpcAdmin)
}

func step3SetupOracle(config *Config) {
	client := config.CreatePClient()
	val, err := client.GetCurrentValidators(context.Background(), ids.Empty, nil)
	panicIfError(err)
	var nodeIds []string
	for _, validator := range val {
		nodeIds = append(nodeIds, validator.NodeID.String())
	}
	valCount, err := config.GetOracle().ValidatorCount(nil)
	panicIfError(err)
	if valCount.Int64() == 0 {
		signer, err := bind.NewKeyedTransactorWithChainID(oracleAndMpcAdmin, config.GetChainId())
		panicIfError(err)
		tx, err := config.GetOracle().StartNodeIDUpdate(signer)
		panicIfError(err)
		waitForTx(config, tx.Hash())
		tx, err = config.GetOracle().AppendNodeIDs(signer, nodeIds)
		panicIfError(err)
		waitForTx(config, tx.Hash())
		tx, err = config.GetOracle().EndNodeIDUpdate(signer)
		panicIfError(err)
		waitForTx(config, tx.Hash())
	}

	setupOracleReport(config)
}

func step4SetupMpc(config *Config) {
	signer, err := bind.NewKeyedTransactorWithChainID(oracleAndMpcAdmin, config.GetChainId())
	panicIfError(err)
	var pubKeyBytes [][]byte
	for _, operator := range mpcOperators {
		bytes := crypto.FromECDSAPub(&operator.PublicKey)
		pubKeyBytes = append(pubKeyBytes, bytes[1:])
	}
	config.GetMpcManager().CreateGroup(signer, pubKeyBytes, uint8(mpcThreshold))
	config.GetMpcManager().RequestKeygen(signer, mpcGroupId)
	genPubKey := crypto.FromECDSAPub(&simMpcPrivKey.PublicKey)
	for i, operator := range mpcOperators {
		doReportKeygen(config, byte(i+1), operator, genPubKey[1:])
	}
}

func panicIfError(err error) {
	if err != nil {
		defaultLogger.Errorf("got err %v", err)
		panic(err)
	}
}

func doFund(c *Config, privKey *ecdsa.PrivateKey) {
	signer, err := bind.NewKeyedTransactorWithChainID(privKey, c.GetChainId())
	panicIfError(err)
	bal, err := c.GetEthClient().BalanceAt(context.Background(), signer.From, nil)
	panicIfError(err)
	if bal.Cmp(tenEthers) > 0 {
		defaultLogger.Debugf("skip funding for %v", signer.From.String())
		fmt.Printf("skip funding for %v\n", signer.From.String())
		return
	}
	signedTx := fundAddrTx(c, signer.From)
	err = c.GetEthClient().SendTransaction(context.Background(), signedTx)
	panicIfError(err)
	txHash := signedTx.Hash()
	waitForTx(c, txHash)
}

func fundAddrTx(c *Config, address common.Address) *types.Transaction {
	signer, err := bind.NewKeyedTransactorWithChainID(deployer, c.GetChainId())
	panicIfError(err)
	nonce, err := c.GetEthClient().NonceAt(context.Background(), signer.From, nil)
	panicIfError(err)
	gasPrice, err := c.GetEthClient().SuggestGasPrice(context.Background())
	panicIfError(err)
	rawTx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &address,
		Value:    tenEthers,
		Gas:      25000,
		GasPrice: gasPrice,
		Data:     nil,
	})
	signedTx, err := signer.Signer(signer.From, rawTx)
	panicIfError(err)
	return signedTx
}

func doSetOracleAddress(config *Config, privKey *ecdsa.PrivateKey, address common.Address) {
	signer, err := bind.NewKeyedTransactorWithChainID(privKey, config.GetChainId())
	panicIfError(err)
	tx, err := config.GetOracleManager().SetOracleAddress(signer, address)
	panicIfError(err)
	waitForTx(config, tx.Hash())
}

func doSetEpochDuration(config *Config, privKey *ecdsa.PrivateKey) {
	signer, err := bind.NewKeyedTransactorWithChainID(privKey, config.GetChainId())
	panicIfError(err)
	tx, err := config.GetOracle().SetEpochDuration(signer, big.NewInt(17))
	panicIfError(err)
	waitForTx(config, tx.Hash())
}

func doReportKeygen(config *Config, ind byte, privKey *ecdsa.PrivateKey, pubKey []byte) {
	signer, err := bind.NewKeyedTransactorWithChainID(privKey, config.GetChainId())
	panicIfError(err)
	var participantId [32]byte
	copy(participantId[:31], mpcGroupId[:31])
	participantId[31] = ind
	_, err = config.GetMpcManager().ReportGeneratedKey(signer, participantId, pubKey)
	panicIfError(err)
}

func setupOracleReport(config *Config) {
	latestEpoch, err := config.GetOracle().LatestFinalizedEpochId(nil)
	panicIfError(err)
	if latestEpoch.Cmp(big.NewInt(0)) > 0 {
		// We only need one finalized report. If there is one, no need to report again.
		return
	}
	reportable, err := config.GetOracle().CurrentReportableEpoch(nil)
	panicIfError(err)
	report := prepareReport(config)
	for _, operator := range oracleOperators[:2] {
		doOracleReport(config, reportable, report, operator)
	}
}

func prepareReport(config *Config) []*big.Int {
	count, err := config.GetOracle().ValidatorCount(nil)
	panicIfError(err)
	allGoodReport := big.NewInt(1023)
	var report []*big.Int
	for i := 0; i < int(count.Int64()); i++ {
		ind := big.NewInt(int64(i))
		ind.Lsh(ind, 10)
		valRep := big.NewInt(0)
		valRep.Or(ind, allGoodReport)
		report = append(report, valRep)
	}
	return report
}

func doOracleReport(config *Config, epoch *big.Int, report []*big.Int, privKey *ecdsa.PrivateKey) {
	signer, err := bind.NewKeyedTransactorWithChainID(privKey, config.GetChainId())
	panicIfError(err)
	signer.GasPrice = big.NewInt(25000000000)
	signer.GasLimit = 900000
	fmt.Printf("sent from %x\n", signer.From)
	tx, err := config.GetOracleManager().ReceiveMemberReport(signer, epoch, report)
	panicIfError(err)
	waitForTx(config, tx.Hash())
}

func waitForTx(config *Config, txHash common.Hash) {
	err := backoff.RetryFnConstant(defaultLogger, context.Background(), 100, 1*time.Second, func() (bool, error) {
		receipt, _ := config.GetEthClient().TransactionReceipt(context.Background(), txHash)
		retry := receipt == nil
		return retry, nil
	})
	panicIfError(err)
}
