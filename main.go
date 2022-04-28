package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"
	"github.com/avalido/mpc-controller/logger"
	mpcTask "github.com/avalido/mpc-controller/task"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"

	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/core"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

const (
	AVAX_ID                 = "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
	CCHAIN_ID               = "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU"
	PROG_NAME               = "mpc-controller"
	PARAM_URL               = "rpc-url"
	PARAM_MPC_SERVICE_URL   = "mpc-url"
	PARAM_COORDINATOR_ADDR  = "coordinator-address"
	PARAM_PRIVATE_KEY       = "private-key"
	ADDR_CCHAIN             = "0x8db97c7cece249c2b98bdc0226cc4c2a57bf52fc"
	ADDR_PCHAIN             = "P-local18jma8ppw3nhx5r4ap8clazz0dps7rv5u00z96u"
	ADDR_CONTRACT           = "0x5aa01B3b5877255cE50cc55e8986a7a5fe29C70e"
	EVENT_PARTICIPANT_ADDED = "ParticipantAdded"
	NETWORK_ID              = 12345
	CHAIN_ID                = 43112
	DELEGATION_FEE          = 100
	GAS_PER_BYTE            = 1
	GAS_PER_SIG             = 1000
	GAS_FIXED               = 10000
	IMPORT_FEE              = 1000000
)

var cChainAddress = common.HexToAddress(ADDR_CCHAIN)
var testnetKey = "56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027"
var keys = []string{
	"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
	"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
	"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
}
var pKeys = []string{
	"c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66defafb10f0932dc5c649debab82e6c816e164c7b7ad8abbe974d15a94cd1c2937",
	"d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e5f4d4ffc52db1898fe83742a8732e53322c178acb7113072c8dc6f82bbc00b99",
	"73ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6f06766bde6ff723360cee7f89abab428717f811f460ebf67f5186f75a9f4288d",
}
var contractAddr = common.HexToAddress(ADDR_CONTRACT)

var abiStr = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"groupId","type":"bytes32"},{"indexed":false,"internalType":"bytes","name":"publicKey","type":"bytes"}],"name":"KeyGenerated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"groupId","type":"bytes32"}],"name":"KeygenRequestAdded","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes","name":"publicKey","type":"bytes"},{"indexed":false,"internalType":"bytes32","name":"groupId","type":"bytes32"},{"indexed":false,"internalType":"uint256","name":"index","type":"uint256"}],"name":"ParticipantAdded","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"requestId","type":"uint256"},{"indexed":true,"internalType":"bytes","name":"publicKey","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"message","type":"bytes"}],"name":"SignRequestAdded","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"requestId","type":"uint256"},{"indexed":true,"internalType":"bytes","name":"publicKey","type":"bytes"},{"indexed":false,"internalType":"bytes","name":"message","type":"bytes"}],"name":"SignRequestStarted","type":"event"},{"inputs":[{"internalType":"bytes[]","name":"publicKeys","type":"bytes[]"},{"internalType":"uint256","name":"threshold","type":"uint256"}],"name":"createGroup","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"groupId","type":"bytes32"}],"name":"getGroup","outputs":[{"internalType":"bytes[]","name":"participants","type":"bytes[]"},{"internalType":"uint256","name":"threshold","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes","name":"publicKey","type":"bytes"}],"name":"getKey","outputs":[{"components":[{"internalType":"bytes32","name":"groupId","type":"bytes32"},{"internalType":"bool","name":"confirmed","type":"bool"}],"internalType":"structMpcCoordinator.KeyInfo","name":"keyInfo","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"requestId","type":"uint256"},{"internalType":"uint256","name":"myIndex","type":"uint256"}],"name":"joinSign","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"groupId","type":"bytes32"},{"internalType":"uint256","name":"myIndex","type":"uint256"},{"internalType":"bytes","name":"generatedPublicKey","type":"bytes"}],"name":"reportGeneratedKey","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"groupId","type":"bytes32"}],"name":"requestKeygen","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes","name":"publicKey","type":"bytes"},{"internalType":"bytes","name":"message","type":"bytes"}],"name":"requestSign","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

func getSigners(key string) ([]*avaCrypto.PrivateKeySECP256K1R, error) {
	f := avaCrypto.FactorySECP256K1R{}
	var signers0 []*avaCrypto.PrivateKeySECP256K1R
	signer, err := f.ToPrivateKey(common.Hex2Bytes(key))
	if err != nil {
		return nil, err
	}
	signers0 = append(signers0, signer.(*avaCrypto.PrivateKeySECP256K1R))
	return signers0, nil
}

func nAVAX(avax uint64) uint64 {
	return avax * 1000_000_000
}

func testNetworkContext() (*core.NetworkContext, error) {
	cchainID, err := ids.FromString(CCHAIN_ID)
	if err != nil {
		return nil, err
	}
	assetId, err := ids.FromString(AVAX_ID)
	if err != nil {
		return nil, err
	}

	ctx := core.NewNetworkContext(
		NETWORK_ID,
		cchainID,
		big.NewInt(CHAIN_ID),
		avax.Asset{
			ID: assetId,
		},
		IMPORT_FEE,
		GAS_PER_BYTE,
		GAS_PER_SIG,
		GAS_FIXED,
	)
	return &ctx, nil
}

func testFlow() error {
	networkCtx, err := testNetworkContext()
	if err != nil {
		return err
	}

	signers0, err := getSigners(testnetKey)
	if err != nil {
		return err
	}

	nodeID, err := ids.ShortFromPrefixedString("NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5", constants.NodeIDPrefix)

	if err != nil {
		return err
	}

	pubkey := signers0[0].ToECDSA().PublicKey

	client, err := ethclient.Dial("http://localhost:9650/ext/bc/C/rpc")

	nonce, err := client.NonceAt(context.Background(), cChainAddress, nil)

	fiveMins := uint64(5 * 60)
	twentyOneDays := uint64(21 * 24 * 60 * 60)
	startTime := uint64(time.Now().Unix()) + fiveMins
	endTime := startTime + twentyOneDays

	task, err := mpcTask.NewStakeTask(*networkCtx, pubkey, nonce, nodeID, nAVAX(40), startTime, endTime, 500)
	if err != nil {
		return err
	}

	hash1, err := task.ExportTxHash()
	if err != nil {
		return err
	}

	sigs := make([][65]byte, 3)

	sig1, err := signers0[0].SignHash(hash1)
	copy(sigs[0][:], sig1[:])

	err = task.SetExportTxSig(sigs[0])
	if err != nil {
		return err
	}

	hash2, err := task.ImportTxHash()
	if err != nil {
		return err
	}

	sig2, err := signers0[0].SignHash(hash2)
	copy(sigs[1][:], sig2[:])
	err = task.SetImportTxSig(sigs[1])
	if err != nil {
		return err
	}

	hash3, err := task.AddDelegatorTxHash()
	if err != nil {
		return err
	}

	sig3, err := signers0[0].SignHash(hash3)
	copy(sigs[2][:], sig3[:])
	err = task.SetAddDelegatorTxSig(sigs[2])
	if err != nil {
		return err
	}

	tx1, err := task.GetSignedExportTx()
	if err != nil {
		return err
	}

	cclient := evm.NewClient("http://localhost:9650", "C")
	txId1, err := cclient.IssueTx(context.Background(), tx1.Bytes())

	if err != nil {
		return err
	}
	fmt.Printf("ExportTx %v\n", txId1)
	time.Sleep(time.Second * 2)
	pclient := platformvm.NewClient("http://localhost:9650")
	tx2, err := task.GetSignedImportTx()
	if err != nil {
		return err
	}
	txId2, err := pclient.IssueTx(context.Background(), tx2.Bytes())
	if err != nil {
		return err
	}
	fmt.Printf("ImportTx %v\n", txId2)
	time.Sleep(time.Second * 2)
	tx3, err := task.GetSignedAddDelegatorTx()
	if err != nil {
		return err
	}

	txId3, err := pclient.IssueTx(context.Background(), tx3.Bytes())
	if err != nil {
		return err
	}
	fmt.Printf("AddDelegatorTx %v\n", txId3)

	return nil
}

type NullMpcClient struct{}

func (n NullMpcClient) Keygen(request *core.KeygenRequest) error {
	return nil
}

func (n NullMpcClient) Sign(request *core.SignRequest) error {
	return nil
}

func (n NullMpcClient) CheckResult(requestId string) (*core.Result, error) {
	return nil, nil
}

func testManager(c *cli.Context) error {

	networkCtx, err := testNetworkContext()
	if err != nil {
		return err
	}

	sk, err := crypto.HexToECDSA(c.String(PARAM_PRIVATE_KEY))
	if err != nil {
		return err
	}

	mpcClient, err := core.NewMpcClient(c.String(PARAM_MPC_SERVICE_URL))
	if err != nil {
		return err
	}
	coordinatorAddr := common.HexToAddress(c.String(PARAM_COORDINATOR_ADDR))
	manager, err := mpcTask.NewTaskManager(
		*networkCtx, mpcClient, sk, coordinatorAddr,
	)
	err = manager.Initialize()
	if err != nil {
		return err
	}
	manager.Start()
	return nil
}

func handler(c *cli.Context) error {

	//err := testFlow()
	err := testManager(c)

	if err != nil {
		panic(err)
	}

	return nil
}

func main() {
	logger.DevMode = true // todo: enable DevMode configurability and delete this line
	app := &cli.App{
		Name:  PROG_NAME,
		Usage: "Handles the MPC operations needed for Avalanche",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     PARAM_URL,
				Required: true,
				Usage:    "The avalanche rpc url.",
			},
			&cli.StringFlag{
				Name:     PARAM_MPC_SERVICE_URL,
				Required: true,
				Usage:    "The mpc url.",
			},
			&cli.StringFlag{
				Name:     PARAM_PRIVATE_KEY,
				Required: true,
				Usage:    "The private key.",
			},
			&cli.StringFlag{
				Name:     PARAM_COORDINATOR_ADDR,
				Required: true,
				Usage:    "The contract address of coordinator.",
			},
		},
		Action: handler,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
