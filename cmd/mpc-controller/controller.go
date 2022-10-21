package main

import (
	"context"
	"fmt"
	"github.com/alitto/pond"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/chain"
	"github.com/avalido/mpc-controller/chain/txissuer"
	"github.com/avalido/mpc-controller/chain/utxotracker"
	"github.com/avalido/mpc-controller/config"
	"github.com/avalido/mpc-controller/contract/caller"
	"github.com/avalido/mpc-controller/contract/transactor"
	"github.com/avalido/mpc-controller/contract/watcher"
	"github.com/avalido/mpc-controller/core"
	"github.com/avalido/mpc-controller/events"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/logger/adapter"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/storage"
	c2pChainStake "github.com/avalido/mpc-controller/tasks/atomicTask/c2pChain/stake"
	p2cChainRecover "github.com/avalido/mpc-controller/tasks/atomicTask/p2cChain/recover"
	joinTaskRecover "github.com/avalido/mpc-controller/tasks/joinTask/recover"
	joinTaskStake "github.com/avalido/mpc-controller/tasks/joinTask/stake"
	pChainTaskStake "github.com/avalido/mpc-controller/tasks/pChainTask/stake"
	"github.com/avalido/mpc-controller/utils/addrs"
	"github.com/avalido/mpc-controller/utils/bytes"
	myCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/keystore"
	"github.com/avalido/mpc-controller/utils/dispatcher" // todo: apply kubecost/events instead
	"github.com/avalido/mpc-controller/utils/noncer"
	"github.com/dgraph-io/ristretto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	kbcevents "github.com/kubecost/events"
	"golang.org/x/sync/errgroup"
	"math/big"
	"reflect"

	"github.com/avalido/mpc-controller/storage/badgerDB"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type Service interface {
	Start() error
	Close() error
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
			if err := service.Start(); err != nil {
				return errors.Wrapf(err, "failed starting service: %v", name)
			}
			//c.Logger.Info(fmt.Sprintf("%v service %v stopped", c.ID, name)) // todo: use CLOSE()
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
	logger.UseConsoleEncoder = config.UseConsoleEncoder
	myLogger := logger.Default()

	// Create nonce manager
	noncer := noncer.New(1, 0) // todo: config it

	// Create dispatcher
	myDispatcher := dispatcher.NewDispatcher(ctx, myLogger, config.ControllerId+"_global", 1024)

	// Create database
	badgerDBLogger := &adapter.BadgerDBLoggerAdapter{Logger: logger.DefaultWithCallerSkip(3)}
	myDB := &badgerDB.BadgerDB{
		Logger: myLogger,
		DB:     badgerDB.NewBadgerDBWithDefaultOptions(config.BadgerDbPath, badgerDBLogger),
	}

	// Get MpcManager contract address
	contractAddr := common.HexToAddress(config.MpcManagerAddress)
	myLogger.Info(fmt.Sprintf("MpcManager address: %v", config.MpcManagerAddress))

	// Create mpcClient
	mpcClient := &core.MyMpcClient{
		Logger:       myLogger,
		MpcServerUrl: config.MpcServerUrl}

	// Decrypt private key
	config.ControllerKey = decryptKey(config.ControllerId, pss, config.ControllerKey)

	// Parse private key
	myPrivKey, err := crypto.HexToECDSA(config.ControllerKey)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to parse private key %q", config.ControllerKey))
	}

	myAddr := addrs.PubkeyToAddresse(&myPrivKey.PublicKey)
	myLogger.Info(fmt.Sprintf("%v address: %v", config.ControllerId, myAddr))

	// Parse public key
	myPubKeyBytes := myCrypto.MarshalPubkey(&myPrivKey.PublicKey)[1:]
	myPartiPubKey := storage.PubKey(myPubKeyBytes)

	// Convert chain ID
	chainId := big.NewInt(config.ChainId)

	// Create transaction signer
	signer, err := bind.NewKeyedTransactorWithChainID(myPrivKey, chainId)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to create transaction signer", logger.Field{"error", err}))
	}

	// Create eth rpc client
	ethRpcClient, err := ethclient.Dial(config.EthRpcUrl)
	if err != nil {
		panic(errors.Wrapf(err, "Failed to connect eth rpc client, url: %q", config.EthRpcUrl))
	}
	rpcEthCliWrapper := &chain.RpcEthClientWrapper{myLogger, ethRpcClient}

	// Create C-Chain issue client
	cChainIssueCli := evm.NewClient(config.CChainIssueUrl, "C")

	// Create P-Chain issue client
	pChainIssueCli := platformvm.NewClient(config.PChainIssueUrl)

	// Create tx issuer
	myTxIssuer := txissuer.MyTxIssuer{
		Logger:       myLogger,
		CChainClient: cChainIssueCli,
		PChainClient: pChainIssueCli,
	}

	boundCaller := caller.MyCaller{
		ContractAddr:   contractAddr,
		ContractCaller: rpcEthCliWrapper,
		Logger:         myLogger,
	}
	boundCaller.Init(ctx)

	boundTransactor := transactor.MyTransactor{
		Auth:               signer,
		ContractAddr:       contractAddr,
		ContractTransactor: rpcEthCliWrapper,
		EthClient:          rpcEthCliWrapper,
		Logger:             myLogger,
	}
	boundTransactor.Init(ctx)

	// Create global dispatcher
	stakeReqAddedDispatcher := kbcevents.GlobalDispatcherFor[*events.StakeRequestAdded]()
	requestStartedDispatcher := kbcevents.GlobalDispatcherFor[*events.RequestStarted]()
	stakeAtomicDispatcher := kbcevents.NewDispatcher[*events.StakeAtomicTaskHandled]()
	utxoToRecoverDispatcher := kbcevents.NewDispatcher[*events.UTXOToRecover]()

	// Create global cache
	globalCache, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	joinTaskRecoverTaskCreator := joinTaskRecover.TaskCreator{
		Ctx:         ctx,
		Logger:      myLogger,
		PartiPubKey: myPartiPubKey,
		DB:          myDB,
		MpcClient:   mpcClient,
		TxIssuer:    &myTxIssuer,
		Network:     networkCtx(config),
		Bound:       &boundTransactor,
		Pool:        pond.New(3, 1000),
		Dispatcher:  utxoToRecoverDispatcher,
	}

	joinTaskStakeTaskCreator := joinTaskStake.TaskCreator{
		Ctx:         ctx,
		Logger:      myLogger,
		PartiPubKey: myPartiPubKey,
		DB:          myDB,
		Bound:       &boundTransactor,
		Pool:        pond.New(3, 1000),
		Dispatcher:  stakeReqAddedDispatcher,
	}

	c2pChainStakeTaskCreator := c2pChainStake.TaskCreator{
		Ctx:                      ctx,
		Logger:                   myLogger,
		MpcClient:                mpcClient,
		TxIssuer:                 &myTxIssuer,
		Network:                  networkCtx(config),
		NonceGiver:               noncer,
		Pool:                     pond.New(3, 1000),
		ReqStartedEvtDispatcher:  requestStartedDispatcher,
		StakeAtomicEvtDispatcher: stakeAtomicDispatcher,
	}

	pChainStakeTaskCreator := pChainTaskStake.TaskCreator{
		Ctx:        ctx,
		Logger:     myLogger,
		MpcClient:  mpcClient,
		TxIssuer:   &myTxIssuer,
		NonceGiver: noncer,
		Network:    networkCtx(config),
		Pool:       pond.New(3, 1000),
		Dispatcher: stakeAtomicDispatcher,
	}

	p2cChainRecoverTaskCreator := p2cChainRecover.TaskCreator{
		Ctx:            ctx,
		Logger:         myLogger,
		MpcClient:      mpcClient,
		TxIssuer:       &myTxIssuer,
		Network:        networkCtx(config),
		ContractCaller: &boundCaller,
		Pool:           pond.New(3, 1000),
		Dispatcher:     requestStartedDispatcher,
		Cache:          globalCache,
	}

	watcherMaster := watcher.Master{
		Ctx:                     ctx,
		BoundCaller:             &boundCaller,
		BoundTransactor:         &boundTransactor,
		ContractAddr:            contractAddr,
		DB:                      myDB,
		EthWsURL:                config.EthWsUrl,
		MpcClient:               mpcClient,
		Logger:                  myLogger,
		PartiPubKey:             myPartiPubKey,
		Dispatcher:              myDispatcher,
		StakeReqAddedDispatcher: stakeReqAddedDispatcher,
		ReqStartedDispatcher:    requestStartedDispatcher,
	}

	utxoTrackerMaster := utxotracker.Master{
		Ctx:                     ctx,
		Logger:                  myLogger,
		Cache:                   globalCache,
		ClientPChain:            pChainIssueCli,
		Dispatcher:              myDispatcher,
		UTXOToRecoverDispatcher: utxoToRecoverDispatcher,
	}

	metricsService := prom.MetricsService{
		Ctx:       ctx,
		ServeAddr: config.MetricsServeAddr,
	}

	controller := MpcController{
		Logger: myLogger,
		ID:     config.ControllerId,
		Services: []Service{
			&watcherMaster,
			&joinTaskRecoverTaskCreator,
			&joinTaskStakeTaskCreator,
			&c2pChainStakeTaskCreator,
			&pChainStakeTaskCreator,
			&utxoTrackerMaster,
			&p2cChainRecoverTaskCreator,
			&metricsService,
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

func decryptKey(id, pss, cipherKey string) string {
	keyBytes, err := keystore.Decrypt(pss, bytes.HexToBytes(cipherKey))
	if err != nil {
		err = errors.Wrapf(err, "%q failed to decrypt  key %q", id, cipherKey)
		panic(fmt.Sprintf("%+v", err))
	}

	var privKey string
	switch len(cipherKey) {
	case 192:
		privKey = string(keyBytes)
	case 128:
		privKey = bytes.BytesToHex(keyBytes)
	}

	return privKey
}
