package core

import (
	"fmt"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type Config struct {
	Host              string
	Port              int16
	SslEnabled        bool
	MpcManagerAddress common.Address
	NetworkContext    chain.NetworkContext
}

func (c Config) getUri() string {
	scheme := "http"
	if c.SslEnabled {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v:%v", scheme, c.Host, c.Port)
}

func (c Config) CreateEthClient() *ethclient.Client {
	client, err := ethclient.Dial(fmt.Sprintf("%s/ext/bc/C/rpc", c.getUri()))
	if err != nil {
		panic(errors.Wrap(err, "failed to get eth client"))
	}
	return client
}

type ServicePack struct {
	Config    Config
	Logger    logger.Logger
	MpcClient MpcClient
	Db        storage.SlimDb
}

func NewServicePack(config Config, logger logger.Logger, mpcClient MpcClient, db storage.SlimDb) *ServicePack {
	return &ServicePack{
		Config:    config,
		Logger:    logger,
		MpcClient: mpcClient,
		Db:        db,
	}
}
