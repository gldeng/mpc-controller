package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/avalido/mpc-controller/core"
	types2 "github.com/avalido/mpc-controller/core/types"
	"github.com/avalido/mpc-controller/eventhandlercontext"
	"github.com/avalido/mpc-controller/logger"
	"github.com/avalido/mpc-controller/mpcclient"
	"github.com/avalido/mpc-controller/pool"
	"github.com/avalido/mpc-controller/router"
	"github.com/avalido/mpc-controller/storage"
	"github.com/avalido/mpc-controller/subscriber"
	"github.com/avalido/mpc-controller/synchronizer"
	"github.com/avalido/mpc-controller/taskcontext"
	"github.com/avalido/mpc-controller/tasks/ethlog"
	"github.com/avalido/mpc-controller/tasks/stake"
	"github.com/avalido/mpc-controller/utils/bytes"
	utilsCrypto "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/avalido/mpc-controller/utils/crypto/keystore"
	"github.com/avalido/mpc-controller/utils/testingutils"
	"github.com/enriquebris/goconcurrentqueue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"math/big"
	"os"
	"os/signal"
	"syscall"
)

// TODO: improve flags

const (
	fnHost              = "host"
	fnPort              = "port"
	fnMpcManagerAddress = "mpc-manager-address"
	fnPrivateKey        = "private-key"
	fnPassword          = "password"
	fnMpcServerUrl      = "mpcServerUrl"
)

func printLog(event interface{}) {
	evt, ok := event.(types.Log)
	if !ok {
		return
	}
	fmt.Printf("Received event log %v\n", evt)
}

func decryptKey(pss, cipherKey string) string {
	keyBytes, err := keystore.Decrypt(pss, bytes.HexToBytes(cipherKey))
	if err != nil {
		err = errors.Wrapf(err, "failed to decrypt key %q", cipherKey)
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
	logger.UseConsoleEncoder = true // temporally for easier debug only
	myLogger := logger.Default()

	shutdownCtx, shutdown := context.WithCancel(context.Background())
	q := goconcurrentqueue.NewFIFO()

	mpcManagerAddr := common.HexToAddress(c.String(fnMpcManagerAddress))

	// Decrypt and parse private key
	privKey := decryptKey(c.String(fnPassword), c.String(fnPrivateKey))
	myPrivKey, err := crypto.HexToECDSA(privKey)
	if err != nil {
		panic("failed to parse private key")
	}

	// Parse public key
	myPubKeyBytes := utilsCrypto.MarshalPubkey(&myPrivKey.PublicKey)[1:]
	//myPartiPubKey := storage.PubKey(myPubKeyBytes)

	// Convert chain ID
	chainId := big.NewInt(43112)

	// Create transaction signer
	signer, err := bind.NewKeyedTransactorWithChainID(myPrivKey, chainId)
	if err != nil {
		panic("failed to create tx signer")
	}

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

	sub, err := subscriber.NewSubscriber(shutdownCtx, myLogger, coreConfig, q)

	db := storage.NewInMemoryDb() // TODO: use persistent db

	// Create mpcClient
	mpcClient := &mpcclient.MyMpcClient{
		Logger:       myLogger,
		MpcServerUrl: c.String(fnMpcServerUrl)}

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

	//ts.enqueueMessages()
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		shutdown()
		rt.Close()
		sub.Close()
		wp.Close()
	}()

	_ = syn.Start()

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
				Name:     fnPrivateKey,
				Required: true,
				Usage:    "The private key for this participant.",
			},
			&cli.StringFlag{
				Name:     fnPassword,
				Required: true,
				Usage:    "The password to decrypt private key",
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
