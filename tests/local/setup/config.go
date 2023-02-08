package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/contract"
	"github.com/avalido/mpc-controller/tests/local/bindings/avalido"
	"github.com/avalido/mpc-controller/tests/local/bindings/oracle"
	"github.com/avalido/mpc-controller/tests/local/bindings/oraclemanager"
	"github.com/avalido/mpc-controller/utils/bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
)

var (
	deployer          *ecdsa.PrivateKey
	oracleAndMpcAdmin *ecdsa.PrivateKey
	oracleOperators   []*ecdsa.PrivateKey
	mpcOperators      []*ecdsa.PrivateKey
	mpcThreshold      = 4
	mpcGroupId        [32]byte
	simMpcPrivKey     *ecdsa.PrivateKey
)

func init() {
	sk, err := crypto.HexToECDSA("56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027")
	panicIfError(err)
	deployer = sk
	simMpcPrivKey = sk
	sk, err = crypto.HexToECDSA("a87518b3691061b9de6dd281d2dda06a4fe3a2c1b4621ac1e05d9026f73065bd")
	panicIfError(err)
	oracleAndMpcAdmin = sk

	mpcGroupIdBytes := bytes.HexTo32Bytes("c9dfdfccdc1a33434ea6494da21cc1e2b03477740c606f0311d1f90665070400")
	copy(mpcGroupId[:], mpcGroupIdBytes[:])
	mpcGroup := []string{
		"353fb105bbf9c29cbf46d4c93a69587ac478138b7715f0786d7ae1cc05230878",
		"7084300e7059ea4b308ec5b965ef581d3f9c9cd63714082ccf9b9d1fb34d658b",
		"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
		"156177364ae1ca503767382c1b910463af75371856e90202cb0d706cdce53c33",
		"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
		"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
		"b17eac91d7aa2bd5fa72916b6c8a35ab06e8f0c325c98067bbc9645b85ce789f",
	}
	for _, sk := range mpcGroup {
		k, err := crypto.HexToECDSA(sk)
		panicIfError(err)
		mpcOperators = append(mpcOperators, k)
	}
	oracleGroup := []string{
		"a54a5d692d239287e8358f27caee92ab5756c0276a6db0a062709cd86451a855",
		"86a5e025e16a96e2706d72fd6115f2ee9ae1c5dfc4c53894b70b19e6fc73b838",
		"d876abc4ef78972fc733651bfc79676d9a6722626f9980e2db249c22ed57dbb2",
		"6353637e9d5cdc0cbc921dadfcc8877d54c0a05b434a1d568423cb918d582eac",
		"c847f461acdd47f2f0bf08b7480d68f940c97bbc6c0a5a03e0cbefae4d9a7592",
	}
	for _, sk := range oracleGroup {
		k, err := crypto.HexToECDSA(sk)
		panicIfError(err)
		oracleOperators = append(oracleOperators, k)
	}
}

type Config struct {
	Host                 string
	Port                 int16
	AvalidoAddress       common.Address
	OracleManagerAddress common.Address
	OracleAddress        common.Address
	ethClient            *ethclient.Client
	avalido              *avalido.AvaLido
	oracleManager        *oraclemanager.OracleManager
	oracle               *oracle.Oracle
	mpcManager           *contract.MpcManager
	chainId              *big.Int
}

func NewConfig(host string, port int16, avalidoAddress, oracleManagerAddress common.Address, oracleAddress common.Address) Config {
	return Config{
		Host:                 host,
		Port:                 port,
		AvalidoAddress:       avalidoAddress,
		OracleManagerAddress: oracleManagerAddress,
		OracleAddress:        oracleAddress,
		ethClient:            nil,
		avalido:              nil,
		oracleManager:        nil,
		oracle:               nil,
		mpcManager:           nil,
		chainId:              nil,
	}
}

func (c Config) getUri() string {
	scheme := "http"
	return fmt.Sprintf("%v://%v:%v", scheme, c.Host, c.Port)
}

func (c Config) getWsUri() string {
	scheme := "ws"
	return fmt.Sprintf("%v://%v:%v", scheme, c.Host, c.Port)
}

func (c Config) CreateWsClient() (*ethclient.Client, error) {
	client, err := ethclient.Dial(fmt.Sprintf("%s/ext/bc/C/ws", c.getWsUri()))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return client, nil
}

func (c Config) GetEthClient() *ethclient.Client {
	if c.ethClient == nil {
		client, err := ethclient.Dial(fmt.Sprintf("%s/ext/bc/C/rpc", c.getUri()))
		panicIfError(err)
		c.ethClient = client
	}

	return c.ethClient
}

func (c Config) CreateCClient() evm.Client {
	return evm.NewClient(c.getUri(), "C")
}

func (c Config) CreatePClient() platformvm.Client {
	return platformvm.NewClient(c.getUri())
}

func (c Config) GetAvaLido() *avalido.AvaLido {
	if c.avalido == nil {
		ethClient := c.GetEthClient()
		inst, err := avalido.NewAvaLido(c.AvalidoAddress, ethClient)
		panicIfError(err)
		c.avalido = inst
	}
	return c.avalido
}

func (c Config) GetOracleManager() *oraclemanager.OracleManager {
	if c.oracleManager == nil {
		ethClient := c.GetEthClient()
		inst, err := oraclemanager.NewOracleManager(c.OracleManagerAddress, ethClient)
		panicIfError(err)
		c.oracleManager = inst
	}
	return c.oracleManager
}

func (c Config) GetMpcManagerAddress() common.Address {
	a := c.GetAvaLido()
	addr, err := a.MpcManager(nil)
	panicIfError(err)
	return addr
}

func (c Config) GetMpcManager() *contract.MpcManager {
	if c.avalido == nil {
		ethClient := c.GetEthClient()
		inst, err := contract.NewMpcManager(c.GetMpcManagerAddress(), ethClient)
		panicIfError(err)
		c.mpcManager = inst
	}
	return c.mpcManager
}

func (c Config) GetOracleAddress() common.Address {
	return c.OracleAddress
}

func (c Config) GetOracle() *oracle.Oracle {
	if c.oracle == nil {
		ethClient := c.GetEthClient()
		inst, err := oracle.NewOracle(c.GetOracleAddress(), ethClient)
		panicIfError(err)
		c.oracle = inst
	}
	return c.oracle
}

func (c Config) GetChainId() *big.Int {
	if c.chainId == nil {
		chainId, err := c.GetEthClient().ChainID(context.Background())
		panicIfError(err)
		c.chainId = chainId
	}
	return c.chainId
}
