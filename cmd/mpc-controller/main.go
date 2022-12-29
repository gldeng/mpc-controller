package main

import (
	"context"
	"fmt"
	"github.com/avalido/mpc-controller/core/mpc"
	"github.com/avalido/mpc-controller/utils/crypto/keystore"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/core"
	types2 "github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/eventhandlercontext"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/prom"
	"github.com/avalido/mpc-controller/router"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/subscriber"
	"github.com/avalido/mpc-controller/synchronizer"
	"github.com/avalido/mpc-controller/taskcontext"
	"github.com/avalido/mpc-controller/tasks/ethlog"
	"github.com/avalido/mpc-controller/tasks/stake"
	"github.com/avalido/mpc-controller/utils/address"
	utilsCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/testingutils"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	cli "github.com/urfave/cli/v2"
)

// TODO: improve flags

const (
	fnHost              = "host"
	fnPort              = "port"
	fnMpcManagerAddress = "mpc-manager-address"
	fnPublicKey         = "publicKey"
	fnKeystoreDir       = "keystoreDir"
	fnPasswordFile      = "passwordFile"
	fnMpcServerUrl      = "mpcServerUrl"
	fnMetricsServeAddr  = "metricsServeAddr"
)

func printLog(event interface{}) {
	evt, ok := event.(types.Log)
	if !ok {
		return
	}
	fmt.Printf("Received event log %v\n", evt)
}

type TestSuite struct {
	db           core.Store
	pubKey       []byte
	queue        *goconcurrentqueue.FIFO
	requestCount int
}

func (s *TestSuite) prepareDb() error {
	err := s.addDummyKey()
	if err != nil {
		return err
	}
	err = s.addDummyRequests()
	if err != nil {
		return err
	}
	err = s.addDummyParticipantId()
	if err != nil {
		return err
	}
	return nil
}

func (s *TestSuite) getRequest(reqNo uint64) *stake.Request {
	return &stake.Request{
		ReqNo:     reqNo,
		TxHash:    common.Hash{},
		PubKey:    s.pubKey,
		NodeID:    "NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5",
		Amount:    "999000000000",
		StartTime: 1663315662,
		EndTime:   1694830062,
	}
}

func (s *TestSuite) addDummyRequests() error {
	for i := 0; i < s.requestCount; i++ {
		s.addDummyRequest(uint64(i))
	}
	return nil
}

func (s *TestSuite) addDummyRequest(reqNo uint64) error {
	request := s.getRequest(reqNo)
	hash, _ := request.Hash()
	fmt.Printf("requestHash is %x\n", hash)
	key := []byte("request/")
	key = append(key, hash[:]...)
	reqBytes, err := request.Encode()
	if err != nil {
		return err
	}
	return s.db.Set(context.Background(), key, reqBytes)
}

func (s *TestSuite) addDummyKey() error {
	key := []byte("key/")
	key = append(key, s.pubKey...)
	keyInfo := types2.MpcPublicKey{
		GroupId:            common.Hash{},
		GenPubKey:          nil,
		ParticipantPubKeys: [][]byte{[]byte("a")},
	}
	bytes, err := keyInfo.Encode()
	if err != nil {
		return err
	}
	return s.db.Set(context.Background(), key, bytes)
}

func (s *TestSuite) addDummyParticipantId() error {
	key := []byte("participant_id")
	var id [32]byte
	id[31] = 1
	return s.db.Set(context.Background(), key, id[:])
}

func (s *TestSuite) enqueueMessages() {
	indices := new(big.Int)
	indices.SetString("8000000000000000000000000000000000000000000000000000000000000000", 16)

	for i := 0; i < s.requestCount; i++ {
		req := s.getRequest(uint64(i))
		h, _ := req.Hash()
		evt := testingutils.MakeEventRequestStarted(h, indices)
		s.queue.Enqueue(evt.Raw)
	}
}

func idFromString(str string) ids.ID {
	id, _ := ids.FromString(str)
	return id
}

func runController(c *cli.Context) error {

	logger.DevMode = true
	logger.UseConsoleEncoder = false // temporally for easier debug only
	myLogger := logger.Default()

	shutdownCtx, shutdown := context.WithCancel(context.Background())
	q := goconcurrentqueue.NewFIFO()
	prom.ConfigFIFOQueueMetrics(q)

	mpcManagerAddr := common.HexToAddress(c.String(fnMpcManagerAddress))

	// Parse public key and address
	pubKey, err := utilsCrypto.UnmarshalPubKeyHex(c.String(fnPublicKey))
	if err != nil {
		panic(fmt.Sprintf("failed to parse public key %q, error: %v", c.String(fnPublicKey), err))
	}
	myAddr := address.PubkeyToAddresse(pubKey)
	myPubKeyBytes := utilsCrypto.MarshalPubkey(pubKey)[1:]

	// Create keystore
	myKeyStore, err := keystore.New(*myAddr, c.String(fnPasswordFile), c.String(fnKeystoreDir))
	if err != nil {
		panic(fmt.Sprintf("failed to create keystore, error: %v", err))
	}

	err = myKeyStore.Unlock()
	if err != nil {
		panic(fmt.Sprintf("failed to unlock keystore, error: %v", err))
	}
	defer myKeyStore.Lock()

	myLogger.Info("set mpc account", []logger.Field{
		{"address", *myAddr},
		{"pubKey", c.String(fnPublicKey)}}...)

	// Convert chain ID
	chainId := big.NewInt(43112)

	// Create signer
	signer, err := bind.NewKeyStoreTransactorWithChainID(myKeyStore.EthKeyStore(), *myKeyStore.Account(), chainId)

	coreConfig := core.Config{
		Host:              c.String(fnHost),
		Port:              int16(c.Int(fnPort)),
		SslEnabled:        false, // TODO: Add argument
		MpcManagerAddress: mpcManagerAddr,
		NetworkContext: core.NewNetworkContext(
			1337,
			idFromString("2cRHidGTGMgWSMQXVuyqB86onp69HTtw6qHsoHvMjk9QbvnijH"),
			chainId,
			avax.Asset{
				ID: idFromString("BUuypiq2wyuLMvyhzFXcPyxPMCgSp7eeDohhQRqTChoBjKziC"),
			},
			1000000,
			1000000,
			1,
			1000,
			10000,
			300,
		),
		MyPublicKey:      myPubKeyBytes,
		MyTransactSigner: signer,
	}
	coreConfig.FetchNetworkInfo()

	db := storage.NewInMemoryDb() // TODO: use persistent db

	// Create mpcClient

	conn, err := grpc.Dial(c.String(fnMpcServerUrl), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "failed to dial to mpc server")
	}
	mpcClient := mpc.NewMpcClient(conn)
	services := core.NewServicePack(coreConfig, myLogger, mpcClient, db)

	syn := synchronizer.NewSyncer(services, q)

	//pk, err := hex.DecodeString("27448e78ffa8cdb24cf19be0204ad954b1bdb4db8c51183534c1eecf2ebd094e28644a0982c69420f823dafe7a062dc9fd4d894be33d088fb02e63ab61710ccb")
	//if err != nil {
	//	return err
	//}
	//ts := &TestSuite{
	//	db:           db,
	//	pubKey:       myPubKeyBytes,
	//	queue:        q,
	//	requestCount: 100,
	//}
	//ts.prepareDb()

	ehContext, err := eventhandlercontext.NewEventHandlerContextImp(services)
	if err != nil {
		return err
	}

	sub, err := subscriber.NewSubscriber(shutdownCtx, myLogger, coreConfig, q, ehContext)

	makeContext := func() core.TaskContext {
		ctx, _ := taskcontext.NewTaskContextImp(services) // TODO: Handler error
		return ctx
	}
	wp, err := pool.NewExtendedWorkerPool(3, makeContext)
	if err != nil {
		return err
	}
	rt, _ := router.NewRouter(q, ehContext, wp)
	rt.AddHandler(printLog)

	rc := &ethlog.RequestCreator{}
	rt.AddLogEventHandler(rc)
	err = wp.Start()
	if err != nil {
		return err
	}
	err = sub.Start()
	if err != nil {
		return err
	}
	err = rt.Start()
	if err != nil {
		return err
	}

	metricsService := prom.MetricsService{
		Ctx:       shutdownCtx,
		ServeAddr: c.String(fnMetricsServeAddr),
	}

	//ts.enqueueMessages()
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		shutdown()
		rt.Close()
		sub.Close()
		wp.Close()
		metricsService.Close()
	}()

	_ = syn.Start()

	go func() {
		metricsService.Start()
	}()

	<-shutdownCtx.Done()

	return nil
}

func main() {

	app := &cli.App{
		Name:  "mpc-controller",
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
				Name:     fnMpcManagerAddress,
				Required: true,
				Usage:    "The address of the deployed MpcManager contract.",
			},
			&cli.StringFlag{
				Name:     fnPublicKey,
				Required: true,
				Usage:    "The public key for this participant.",
			},
			&cli.StringFlag{
				Name:     fnKeystoreDir,
				Required: true,
				Usage:    "The keystore directory for this participant.",
			},
			&cli.StringFlag{
				Name:     fnPasswordFile,
				Required: true,
				Usage:    "The password file to decrypt private key",
			},
			&cli.StringFlag{
				Name:     fnMpcServerUrl,
				Required: true,
				Usage:    "The URL of the MpcServer",
			},
			&cli.StringFlag{
				Name:     fnMetricsServeAddr,
				Required: false,
				Usage:    "The URL of Prometheus metrics service",
			},
		},
		Action: runController,
	}

	fmt.Printf("Starting process: %v\n", os.Getpid())

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Failed to run controller, error: %+v", err)
	}
}
