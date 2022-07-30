package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/cache"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/contract/reconnector"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/dispatcher"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/logger/adapter"
	"github.com/avalido/mpc-controller/support/keygen"
	"github.com/avalido/mpc-controller/tasks/rewarding"
	"github.com/avalido/mpc-controller/tasks/staking"
	"github.com/avalido/mpc-controller/tasks/staking/joining"
	"github.com/avalido/mpc-controller/utils/bytes"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/hash256"
	"github.com/avalido/mpc-controller/utils/crypto/keystore"
	"github.com/avalido/mpc-controller/utils/network"
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"
	"math/big"
	"reflect"

	"github.com/avalido/mpc-controller/storage/badgerDB"
	"github.com/avalido/mpc-controller/support/participant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type Service interface {
	Start(ctx context.Context) error
}

type MpcController struct {
	Logger   logger.Logger
	ID       string
	Services []Service
}

func (c *MpcController) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, service := range c.Services {
		name := reflect.TypeOf(service).String()

		service := service
		g.Go(func() error {
			c.Logger.Info(fmt.Sprintf("%v service %v starting...", c.ID, name))
			if err := service.Start(ctx); err != nil {
				return errors.Wrapf(err, "failed starting service: %v", name)
			}
			c.Logger.Info(fmt.Sprintf("%v service %v stopped", c.ID, name))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func NewController(ctx context.Context, c *cli.Context) *MpcController {
	// Parse config
	config := config.ParseFile(c.String(configFile))
	pss := c.String(password)

	// Create logger
	logger.DevMode = config.EnableDevMode
	myLogger := logger.Default()

	// Create nonce manager
	noncer := noncer.New(1, 0) // todo: config it

	// Create database
	badgerDBLogger := &adapter.BadgerDBLoggerAdapter{Logger: logger.DefaultWithCallerSkip(3)}
	myDB := &badgerDB.BadgerDB{
		Logger: myLogger,
		DB:     badgerDB.NewBadgerDBWithDefaultOptions(config.BadgerDbPath, badgerDBLogger),
	}

	// Create dispatcher
	myDispatcher := dispatcher.NewDispatcher(ctx, myLogger, 1024)

	// Get MpcManager contract address
	contractAddr := common.HexToAddress(config.MpcManagerAddress)

	// Create mpcClient
	mpcClient, _ := core.NewMpcClient(myLogger, config.MpcServerUrl, config.ControllerId)

	// Decrypt private key
	config.ControllerKey = decryptKey(config.ControllerId, pss, config.ControllerKey)

	// Parse private myPrivKey
	myPrivKey, err := crypto.HexToECDSA(config.ControllerKey)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to parse controller private key %q", config.ControllerKey))
	}

	// Parse public myPrivKey
	myPubKeyBytes := myCrypto.MarshalPubkey(&myPrivKey.PublicKey)[1:]
	myPubKeyHex := bytes.BytesToHex(myPubKeyBytes)
	myPubKeyHash := hash256.FromBytes(myPubKeyBytes)

	// Convert chain ID
	chainId := big.NewInt(config.ChainId)

	// Create controller transaction signer
	signer, err := bind.NewKeyedTransactorWithChainID(myPrivKey, chainId)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to create controller transaction signer", logger.Field{"error", err}))
	}

	// Create eth rpc client
	ethRpcClient, err := ethclient.Dial(config.EthRpcUrl)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to connect eth rpc client, url: %q", config.EthRpcUrl))
	}
	rpcEthCliWrapper := &chain.RpcEthClientWrapper{myLogger, ethRpcClient}

	// Create eth ws client
	ethWsClient, err := ethclient.Dial(config.EthWsUrl) // todo: use chain.WsEthClientWrapper
	if err != nil {
		panic(errors.Wrapf(err, "Failed to connect eth ws client, url: %q", config.EthWsUrl))
	}

	// Create C-Chain issue client
	cChainIssueCli := evm.NewClient(config.CChainIssueUrl, "C")
	evmClientWrapper := &chain.EvmClientWrapper{myLogger, cChainIssueCli}

	// Create P-Chain issue client
	pChainIssueCli := platformvm.NewClient(config.PChainIssueUrl)
	platformvmClientWrapper := &chain.PlatformvmClientWrapper{myLogger, pChainIssueCli}

	// Create contract filterer reconnector
	filterReconnector := reconnector.ContractFilterReconnector{
		Logger:    myLogger,
		Updater:   &network.EthClientDialerImpl{myLogger, config.EthWsUrl, ethWsClient},
		Publisher: myDispatcher,
	}

	cacheWrapper := cache.CacheWrapper{
		Dispatcher: myDispatcher,
	}

	participantMaster := participant.ParticipantMaster{
		ContractAddr:    contractAddr,
		ContractCaller:  rpcEthCliWrapper,
		Dispatcher:      myDispatcher,
		Logger:          myLogger,
		MyPubKeyBytes:   myPubKeyBytes,
		MyPubKeyHashHex: myPubKeyHash.Hex(),
		MyPubKeyHex:     myPubKeyHex,
		Storer:          myDB,
	}

	keygenMaster := keygen.KeygenMaster{
		ContractAddr:    contractAddr,
		Dispatcher:      myDispatcher,
		KeygenDoner:     mpcClient,
		Logger:          myLogger,
		MyPubKeyHashHex: myPubKeyHash.Hex(),
		Receipter:       rpcEthCliWrapper,
		Signer:          signer,
		Storer:          myDB,
		Transactor:      rpcEthCliWrapper,
	}

	joiningMaster := joining.JoiningMaster{
		ContractAddr:    contractAddr,
		Dispatcher:      myDispatcher,
		Logger:          myLogger,
		MyIndexGetter:   &cacheWrapper,
		MyPubKeyHashHex: myPubKeyHash.Hex(),
		Receipter:       rpcEthCliWrapper,
		Signer:          signer,
		Transactor:      rpcEthCliWrapper,
	}

	stakingMaster := staking.StakingMaster{
		Balancer:          ethRpcClient,
		CChainIssueClient: evmClientWrapper,
		Cache:             &cacheWrapper,
		ChainNoncer:       rpcEthCliWrapper,
		ContractAddr:      contractAddr,
		Dispatcher:        myDispatcher,
		Logger:            myLogger,
		MyPubKeyHashHex:   myPubKeyHash.Hex(),
		NetworkContext:    networkCtx(config),
		Noncer:            noncer,
		PChainIssueClient: platformvmClientWrapper,
		SignDoner:         mpcClient,
	}

	rewardMaster := rewarding.Master{
		CChainIssueClient: evmClientWrapper,
		Cache:             &cacheWrapper,
		ContractAddr:      contractAddr,
		Dispatcher:        myDispatcher,
		Logger:            myLogger,
		MyPubKeyHashHex:   myPubKeyHash.Hex(),
		NetworkContext:    networkCtx(config),
		PChainClient:      platformvmClientWrapper,
		Receipter:         rpcEthCliWrapper,
		RewardUTXOGetter:  platformvmClientWrapper,
		SignDoner:         mpcClient,
		Signer:            signer,
		Transactor:        rpcEthCliWrapper,
	}

	controller := MpcController{
		Logger: myLogger,
		ID:     config.ControllerId,
		Services: []Service{
			&cacheWrapper,
			&filterReconnector,
			&participantMaster,
			&keygenMaster,
			&joiningMaster,
			&stakingMaster,
			&rewardMaster,
		},
	}

	return &controller
}

func networkCtx(config *config.Config) chain.NetworkContext {
	// Convert C-Chain ID
	cchainID, err := ids.FromString(config.CChainId)
	if err != nil {
		panic(errors.Wrap(err, "Failed to convert C-Chain ID"))
	}

	// Convert chain ID
	chainIdBigInt := big.NewInt(config.ChainId)

	// Convert AVAX assetId ID
	assetId, err := ids.FromString(config.AvaxId)
	if err != nil {
		panic(errors.Wrap(err, "Failed to convert AVAX assetId"))
	}

	// Create NetworkContext
	networkCtx := chain.NewNetworkContext(
		config.NetworkId,
		cchainID,
		chainIdBigInt,
		avax.Asset{
			ID: assetId,
		},
		config.ImportFee,
		config.ExportFee,
		config.GasPerByte,
		config.GasPerSig,
		config.GasFixed,
	)
	return networkCtx
}

func decryptKey(id, pss, keyHex string) string {
	keyBytes, err := keystore.Decrypt(pss, bytes.HexToBytes(keyHex))
	if err != nil {
		err = errors.Wrapf(err, "failed to decrypt keyHex for %v", id)
		panic(fmt.Sprintf("%+v", err))
	}
	return string(keyBytes)
}
