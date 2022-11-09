package core

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type Config struct {
	Host              string
	Port              int16
	SslEnabled        bool
	MpcManagerAddress common.Address
	NetworkContext    NetworkContext
	MyPublicKey       []byte
	MyTransactSigner  *bind.TransactOpts
}

func (c Config) getUri() string {
	scheme := "http"
	if c.SslEnabled {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v:%v", scheme, c.Host, c.Port)
}

func (c Config) getWsUri() string {
	scheme := "ws"
	if c.SslEnabled {
		scheme = "wss"
	}
	return fmt.Sprintf("%v://%v:%v", scheme, c.Host, c.Port)
}

func (c Config) CreateWsClient() *ethclient.Client {
	client, err := ethclient.Dial(fmt.Sprintf("%s/ext/bc/C/ws", c.getWsUri()))
	if err != nil {
		panic(errors.Wrap(err, "failed to get eth client"))
	}
	return client
}

func (c Config) CreateEthClient() *ethclient.Client {
	client, err := ethclient.Dial(fmt.Sprintf("%s/ext/bc/C/rpc", c.getUri()))
	if err != nil {
		panic(errors.Wrap(err, "failed to get eth client"))
	}
	return client
}

func (c Config) CreateCClient() evm.Client {
	return evm.NewClient(c.getUri(), "C")
}

func (c Config) CreatePClient() platformvm.Client {
	return platformvm.NewClient(c.getUri())
}

func (c Config) FetchNetworkInfo() {
	ethClient := c.CreateEthClient()
	//networkID, _ := ethClient.NetworkID(context.Background())
	//if networkID != nil {
	//	c.NetworkContext.SetNetworkID(networkID)
	//}
	chainID, _ := ethClient.ChainID(context.Background())
	if chainID != nil {
		c.NetworkContext.SetChainID(chainID)
	}
	pClient := c.CreatePClient()
	chains, _ := pClient.GetBlockchains(context.Background())
	for _, blockchain := range chains {
		if blockchain.Name == "C-Chain" {
			c.NetworkContext.SetCChainID(blockchain.ID)
		}
	}
	assetID, err := pClient.GetStakingAssetID(context.Background(), ids.Empty)
	if err == nil {
		c.NetworkContext.SetAssetID(assetID)
	}
}

type ServicePack struct {
	Config    Config
	Logger    logger.Logger
	MpcClient MpcClient
	Db        Store
}

func NewServicePack(config Config, logger logger.Logger, mpcClient MpcClient, db Store) *ServicePack {
	return &ServicePack{
		Config:    config,
		Logger:    logger,
		MpcClient: mpcClient,
		Db:        db,
	}
}
