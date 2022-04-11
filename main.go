package main

import (
	"context"
	"fmt"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/constants"
	avaCrypto "github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
	avaEthclient "github.com/ava-labs/coreth/ethclient"

	//"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm"

	"github.com/ava-labs/coreth/plugin/evm"
	"github.com/avalido/mpc-controller/contract"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

const (
	AVAX_ID                 = "2fombhL7aGPwj3KH4bfrmJwW6PVnMobf9Y2fn9GwxiAAJyFDbe"
	CCHAIN_ID               = "2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU"
	PROG_NAME               = "mpc-controller"
	PARAM_URL               = "rpc-url"
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

func decodeUTXOs(utxos [][]byte) []*avax.UTXO {
	out := []*avax.UTXO{}
	for _, bytes := range utxos {
		utxo := new(avax.UTXO)
		_, err := platformvm.Codec.Unmarshal(bytes, &utxo)
		if err == nil {
			out = append(out, utxo)
		}
	}
	return out
}

func spend(utxo *avax.UTXO) *avax.TransferableInput {
	return &avax.TransferableInput{
		UTXOID: utxo.UTXOID,
		Asset:  utxo.Asset,
		In: &secp256k1fx.TransferInput{
			Amt: utxo.Out.(*secp256k1fx.TransferOutput).Amt,
			Input: secp256k1fx.Input{
				SigIndices: []uint32{0},
			},
		},
	}
}

func stakeOut(utxo *avax.UTXO) *avax.TransferableOutput {
	return &avax.TransferableOutput{
		Asset: utxo.Asset,
		Out: &secp256k1fx.TransferOutput{
			Amt: utxo.Out.(*secp256k1fx.TransferOutput).Amt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     utxo.Out.(*secp256k1fx.TransferOutput).Addrs,
			},
		},
	}
}

func feeChangeOut(utxo *avax.UTXO, amt uint64) *avax.TransferableOutput {
	return &avax.TransferableOutput{
		Asset: utxo.Asset,
		Out: &secp256k1fx.TransferOutput{
			Amt: amt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     utxo.Out.(*secp256k1fx.TransferOutput).Addrs,
			},
		},
	}
}

func getSigners(key string) ([]*avaCrypto.PrivateKeySECP256K1R, error) {
	f := avaCrypto.FactorySECP256K1R{}
	signers0 := []*avaCrypto.PrivateKeySECP256K1R{}
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

func getPChainAddr() (*ids.ShortID, error) {
	_, _, addrBytes, err := formatting.ParseAddress(ADDR_PCHAIN)

	if err != nil {
		return nil, err
	}

	addr, err := ids.ToShortID(addrBytes)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func testDecodingExport() error {
	client := evm.NewClient("http://localhost:9650", "C")
	expID, err := ids.FromString("RCeazcWzQjWVmapmUyubGHaVfDeo3QnpAnt3NsXiQdwK8QNXK")
	bytes, err := client.GetAtomicTx(context.Background(), expID)
	if err != nil {
		return err
	}
	tx := new(evm.Tx)
	_, err = evm.Codec.Unmarshal(bytes, &tx)
	if err != nil {
		return err
	}
	fmt.Println(tx)
	assetID := tx.UnsignedAtomicTx.(*evm.UnsignedExportTx).ExportedOutputs[0].Asset.ID.String()
	fmt.Println(assetID)
	blockchainID := tx.UnsignedAtomicTx.(*evm.UnsignedExportTx).BlockchainID.String()
	fmt.Println(blockchainID)
	return nil
}

type MyContext struct {
	networkID uint32
	cchainID  ids.ID
	//assetID ids.ID
	asset        avax.Asset
	avaEthclient avaEthclient.Client
	myAddr       ids.ShortID
}

func NewMyContext() (*MyContext, error) {
	cchainID, err := ids.FromString(CCHAIN_ID)
	if err != nil {
		return nil, err
	}
	assetId, err := ids.FromString(AVAX_ID)
	if err != nil {
		return nil, err
	}

	addr, err := getPChainAddr()
	if err != nil {
		return nil, err
	}
	avaEthclient, err := avaEthclient.Dial("http://localhost:9650/ext/bc/C/rpc")
	if err != nil {
		return nil, err
	}
	return &MyContext{
		networkID: 12345,
		cchainID:  cchainID,
		asset: avax.Asset{
			ID: assetId,
		},
		avaEthclient: avaEthclient,
		myAddr:       *addr,
	}, nil
}

func getGas(tx *evm.UnsignedExportTx) uint64 {
	return uint64(len(tx.Bytes())*GAS_PER_BYTE + len(tx.Ins)*GAS_PER_SIG + GAS_FIXED)
}

func (myContext *MyContext) buildExportTx(nonce uint64, avaxAmt uint64, baseFeeGwei uint64) evm.UnsignedExportTx {
	exportAmt := nAVAX(avaxAmt) + IMPORT_FEE
	input := evm.EVMInput{
		Address: cChainAddress,
		Amount:  exportAmt,
		AssetID: myContext.asset.ID,
		Nonce:   nonce,
	}
	var outs []*avax.TransferableOutput
	outs = append(outs, &avax.TransferableOutput{
		Asset: myContext.asset,
		Out: &secp256k1fx.TransferOutput{
			Amt: exportAmt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					myContext.myAddr,
				},
			},
		},
	})

	tx := evm.UnsignedExportTx{
		NetworkID:        NETWORK_ID,
		BlockchainID:     myContext.cchainID,
		DestinationChain: ids.Empty,
		Ins: []evm.EVMInput{
			input,
		},
		ExportedOutputs: outs,
	}

	gas := getGas(&tx)
	exportFee := gas * baseFeeGwei
	tx.Ins[0].Amount += exportFee
	return tx
}

func (myContext *MyContext) paySelf(amt uint64) avax.TransferableOutput {
	return avax.TransferableOutput{
		Asset: myContext.asset,
		Out: &secp256k1fx.TransferOutput{
			Amt: amt,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					myContext.myAddr,
				},
			},
		},
	}
}

func (myContext *MyContext) buildImportTx(expTx *evm.UnsignedExportTx) platformvm.UnsignedImportTx {
	// TODO: Take sum instead of assuming single?
	index := uint32(0)
	amt := expTx.ExportedOutputs[index].Out.Amount()
	utxo := myContext.paySelf(amt - IMPORT_FEE)
	tx := platformvm.UnsignedImportTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    myContext.networkID,
			BlockchainID: ids.Empty,
			Outs: []*avax.TransferableOutput{
				&utxo,
			},
		}},
		SourceChain: myContext.cchainID,
		ImportedInputs: []*avax.TransferableInput{{
			UTXOID: avax.UTXOID{
				TxID:        expTx.ID(),
				OutputIndex: index,
			},
			Asset: myContext.asset,
			In: &secp256k1fx.TransferInput{
				Amt: amt,
				Input: secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
		}},
	}
	return tx
}

func (myContext *MyContext) buildAddDelegatorTx(impTx *platformvm.UnsignedImportTx, nodeID ids.ShortID) platformvm.UnsignedAddDelegatorTx {
	var (
		signersPlaceholder []*avaCrypto.PrivateKeySECP256K1R
		ins                []*avax.TransferableInput
		returnedOuts       []*avax.TransferableOutput
		stakedOuts         []*avax.TransferableOutput
	)

	fiveMins := uint64(5 * 60)
	twentyOneDays := uint64(21 * 24 * 60 * 60)
	startTime := uint64(time.Now().Unix()) + fiveMins
	endTime := startTime + twentyOneDays
	utxos := impTx.UTXOs()
	utxo := utxos[0]

	stakedOuts = append(stakedOuts, stakeOut(utxo))
	ins = append(ins, spend(utxo))
	signers := [][]*avaCrypto.PrivateKeySECP256K1R{
		signersPlaceholder,
	}
	avax.SortTransferableInputsWithSigners(ins, signers)
	tx := platformvm.UnsignedAddDelegatorTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    12345,
			BlockchainID: ids.Empty,
			Ins:          ins,
			Outs:         returnedOuts,
		}},
		Validator: platformvm.Validator{
			NodeID: nodeID,
			Start:  startTime,
			End:    endTime,
			Wght:   utxo.Out.(*secp256k1fx.TransferOutput).Amt,
		},
		Stake:        stakedOuts,
		RewardsOwner: &utxo.Out.(*secp256k1fx.TransferOutput).OutputOwners,
	}
	return tx
}

func (myContext *MyContext) buildStakeTxFromImported(impTx *evm.UnsignedImportTx) *platformvm.UnsignedAddDelegatorTx {
	return nil
}

/*
func testExport() error {
	addr, err := getPChainAddr()

	if err != nil {
		return err
	}

	//pChainAddress, err := ids.ShortFromPrefixedString(ADDR_PCHAIN, "P-local")
	blockchainId, err := ids.FromString("2CA6j5zYzasynPsFeNoqWkmTCt3VScMvXUZHbfDJ8k3oGzAPtU")
	if err != nil {
		return err
	}

	fmt.Println(blockchainId.String())
	assetId, err := ids.FromString(AVAX_ID)
	if err != nil {
		return err
	}
	fmt.Println(assetId.String())
	ethClient, err := avaEthclient.Dial("http://localhost:9650/ext/bc/C/rpc")
	if err != nil {
		return err
	}
	nonce, err := ethClient.NonceAt(context.Background(), cChainAddress, nil)
	if err != nil {
		return err
	}
	//expID, err := ids.FromString("RCeazcWzQjWVmapmUyubGHaVfDeo3QnpAnt3NsXiQdwK8QNXK")
	//bytes, err := client.GetAtomicTx(context.Background(), expID)
	//if err != nil {
	//	return err
	//}

	input := evm.EVMInput{
		Address: cChainAddress,
		Amount:  100,
		AssetID: assetId,
		Nonce:   nonce,
	}
	outs := []*avax.TransferableOutput{}
	outs = append(outs, &avax.TransferableOutput{
		Asset: avax.Asset{
			ID: assetId,
		},
		Out: &secp256k1fx.TransferOutput{
			Amt: 100,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{*addr},
			},
		},
	})

	utx := &evm.UnsignedExportTx{
		NetworkID:        NETWORK_ID,
		BlockchainID:     blockchainId,
		DestinationChain: ids.Empty,
		Ins: []evm.EVMInput{
			input,
		},
		ExportedOutputs: outs,
	}
	tx := &evm.Tx{UnsignedAtomicTx: utx}

	client := evm.NewClient("http://localhost:9650", "C")
	return nil
}
*/

func testStake() ([][]byte, error) {
	signers0, err := getSigners(testnetKey)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	client := platformvm.NewClient("http://localhost:9650")
	address := []string{"P-local18jma8ppw3nhx5r4ap8clazz0dps7rv5u00z96u"}

	res, _, err := client.GetUTXOs(ctx, address, 1024, "", "")
	if err != nil {
		return nil, err
	}
	fmt.Println(res)
	res1, err := client.GetCurrentValidators(ctx, ids.Empty, []ids.ShortID{})

	if err != nil {
		return nil, err
	}
	fmt.Println(res1)
	nodeID, err := ids.ShortFromPrefixedString("NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5", constants.NodeIDPrefix)

	if err != nil {
		return nil, err
	}

	fiveMins := uint64(5 * 60)
	twentyOneDays := uint64(21 * 24 * 60 * 60)
	startTime := uint64(time.Now().Unix()) + fiveMins
	endTime := startTime + twentyOneDays
	utxos := decodeUTXOs(res)
	feeUtxo := utxos[0]
	delUtxo := utxos[1]

	//txres, err := client.GetTx(ctx, delUtxo.TxID)
	//tx := new(platformvm.Tx)
	//_, err = platformvm.Codec.Unmarshal(txres, &tx)

	feeChange := feeUtxo.Out.(*secp256k1fx.TransferOutput).Amt - DELEGATION_FEE
	ins := []*avax.TransferableInput{}
	returnedOuts := []*avax.TransferableOutput{}
	returnedOuts = append(returnedOuts, feeChangeOut(feeUtxo, feeChange))

	stakedOuts := []*avax.TransferableOutput{}
	stakedOuts = append(stakedOuts, stakeOut(delUtxo))
	// full UTXO is used for delegation
	ins = append(ins, spend(delUtxo))
	// append Fee
	ins = append(ins, spend(feeUtxo))
	signers := [][]*avaCrypto.PrivateKeySECP256K1R{
		signers0,
		signers0,
	}
	avax.SortTransferableInputsWithSigners(ins, signers)
	addDel := &platformvm.UnsignedAddDelegatorTx{
		BaseTx: platformvm.BaseTx{BaseTx: avax.BaseTx{
			NetworkID:    12345,
			BlockchainID: ids.Empty,
			Ins:          ins,
			Outs:         returnedOuts,
		}},
		Validator: platformvm.Validator{
			NodeID: nodeID,
			Start:  startTime,
			End:    endTime,
			Wght:   delUtxo.Out.(*secp256k1fx.TransferOutput).Amt,
		},
		Stake:        stakedOuts,
		RewardsOwner: &delUtxo.Out.(*secp256k1fx.TransferOutput).OutputOwners,
	}
	addDel.UnsignedBytes()

	tx := &platformvm.Tx{UnsignedTx: addDel}

	tx.Sign(platformvm.Codec, signers)
	txId, err := client.IssueTx(ctx, tx.Bytes())
	if err != nil {
		return nil, err
	}
	fmt.Println("txid is %v", txId)
	return res, nil
}

func testFilter() {
	client, err := ethclient.Dial("ws://127.0.0.1:9650/ext/bc/C/ws")
	if err != nil {
		fmt.Print("Failed to connect to websocket")
	}
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
	}
	eventLogs := make(chan types.Log)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = client.SubscribeFilterLogs(ctx, query, eventLogs)
	if err != nil {
		fmt.Printf("Error %v\n", err)
	}

	fmt.Print("Testing filter logs")

	go func(eventLogs chan types.Log) {
		fmt.Print("Go routine started")
		for {
			select {
			case evt, ok := <-eventLogs:
				if ok {
					fmt.Println(evt)
				} else {
					break
				}
			case <-time.After(1 * time.Second):
				fmt.Println("Hearbeat")
			}
		}
		fmt.Print("Go routine stopped")
	}(eventLogs)
}

func startFilterer() error {
	wsclient, err := ethclient.Dial("ws://127.0.0.1:9650/ext/bc/C/ws")
	if err != nil {
		return err
	}

	coordinator, err := contract.NewMpcCoordinator(contractAddr, wsclient)
	events := make(chan *contract.MpcCoordinatorParticipantAdded)
	var start = uint64(1)
	opts := new(bind.WatchOpts)
	opts.Start = &start
	_, err = coordinator.WatchParticipantAdded(opts, events, [][]byte{})
	if err != nil {
		return err
	}
	go func(events chan *contract.MpcCoordinatorParticipantAdded) {
		for {
			select {
			case evt, ok := <-events:
				if ok {
					fmt.Println(evt)
				} else {
					break
				}
			}
		}

	}(events)
	return nil
}

func sendTxn() error {
	client, err := ethclient.Dial("http://localhost:9650/ext/bc/C/rpc")
	if err != nil {
		return err
	}
	var transactors [3]*bind.TransactOpts
	var pubKeys [3][]byte

	for i, k := range keys {
		sk, err := crypto.HexToECDSA(k)
		if err != nil {
			return err
		}
		txactor, err := bind.NewKeyedTransactorWithChainID(sk, big.NewInt(CHAIN_ID))
		if err != nil {
			return err
		}
		transactors[i] = txactor
	}

	for i, k := range pKeys {
		pk := common.Hex2BytesFixed(k, 64)
		pubKeys[i] = pk
	}

	coordinator, err := contract.NewMpcCoordinator(contractAddr, client)

	pubKeys[2][0] = pubKeys[2][0] + 3
	txn, err := coordinator.CreateGroup(transactors[0], pubKeys[:], big.NewInt(1))
	if err != nil {
		return err
	}
	fmt.Printf("Transaction is %v\n", txn.Hash())
	return nil
}

func testDecodingCChain() error {
	expTxStr := "0x000000000001000030399d0775f450604bd2fbc49ce0c5c1c6dfeb2dc2acb8c92c26eeae6e6df4502b190000000000000000000000000000000000000000000000000000000000000000000000018db97c7cece249c2b98bdc0226cc4c2a57bf52fc00000005d2325719dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db000000000000000400000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000700000005d22cfc40000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c000000010000000900000001af89ff91cd986a8455a44d0dc1de4a274da8042c1f28451ae7fc421cdd8137a012b62ed7d862ae37e95de9333fdc8a58e7e636de112903e7600b5851f1ebe764010e1834fd"
	str := expTxStr
	bytes, err := formatting.Decode(formatting.Hex, str)
	if err != nil {
		return err
	}
	tx := new(evm.Tx)
	evm.Codec.Unmarshal(bytes, &tx)

	if err != nil {
		return err
	}
	return nil
}

func testDecodingPChain() error {
	//prevTxStr := "0x00000000001100003039000000000000000000000000000000000000000000000000000000000000000000000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000700000019d81d9600000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c0000000000000000d891ad56056d9c01f18f43f58b5c784ad07a4a49cf3d1f11623804b5cba2c6bf00000001a2581c6a372dc909b3cfbac321526b7de3207a17155783eefa818336f7d7acdd00000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000500000019d82cd84000000001000000000000000100000009000000017bb6ffd4f56ed5d66dcbf98c3a5785bc58787419146dd691e44f22647ef0556a0ae237bdcc7979ec4cab9f914a10fa6cd914d1b8c5f20f2cce1ab76fdd589f8600ed6cdfbf"
	//stakeTxStr := "0x00000000000e00003039000000000000000000000000000000000000000000000000000000000000000000000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db00000007000000003b9aca00000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c00000001690cf8454a3929e03d7723cdb149a9b29a62ee482c97f1ca0e7368b2ed6cdfbf00000000dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000500000019d81d9600000000010000000000000000f29bce5f34a74301eb0de716d5194e4a4aea5d7a00000000624e7ed900000000626a3035000000199c82cc0000000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db00000007000000199c82cc00000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c0000000b000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c0000000100000009000000014416d56e38910f2b88c9c31f914b5d32112b528aa45e53d5bc892435703978a867faf5eac419c47f5bed72435916911ff0d80fe730fb8da7cb61f188f4b6308401109cb412"
	//stakeTxStr := "0x00000000000e00003039000000000000000000000000000000000000000000000000000000000000000000000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db00000007000000003b9ac99c000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c00000002b57a513e94623c7e4847ccd538b97d814550fa26e2bcedcf3cd5c70c5ce81e3900000000dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000500000005d21dba000000000100000000e6bf35511caf215351bee8c0399e1a6a496a89f28058cd06971e5b0cb141739200000000dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db00000005000000003b9aca00000000010000000000000000f29bce5f34a74301eb0de716d5194e4a4aea5d7a00000000624f8c0600000000626b3b8600000005d21dba0000000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000700000005d21dba00000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c0000000b000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c00000002000000090000000161eec95ca84c6104fa55495ef9b4733165893367ab787517efcef26ebfeb0b5d3adcd3da0e3264980853c98203d41c1cbf9746fcc9002a740d07f4a92730c55d01000000090000000161eec95ca84c6104fa55495ef9b4733165893367ab787517efcef26ebfeb0b5d3adcd3da0e3264980853c98203d41c1cbf9746fcc9002a740d07f4a92730c55d01a6552f4d"

	impTxStr := "0x00000000001100003039000000000000000000000000000000000000000000000000000000000000000000000001dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000700000005d21dba00000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c00000000000000009d0775f450604bd2fbc49ce0c5c1c6dfeb2dc2acb8c92c26eeae6e6df4502b190000000136f370deadea98dbbb097927bef5901a8d710b93f44a2fc11a8524f90e1834fd00000000dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db0000000500000005d22cfc400000000100000000000000010000000900000001c8b3f0c23dfa4f0446019093822e6a09257e3af0d915ccbcb33a2a05f245dbb555c801369abdb63cdd441dec5ecbd19531ec77680b06ce950f477fcfca4c14a4015ce81e39"
	//str := prevTxStr
	//str := stakeTxStr
	str := impTxStr
	bytes, err := formatting.Decode(formatting.Hex, str)
	if err != nil {
		return err
	}
	tx := new(platformvm.Tx)
	_, err = platformvm.Codec.Unmarshal(bytes, &tx)

	if err != nil {
		return err
	}
	//impTx := tx.UnsignedTx.(*platformvm.UnsignedImportTx)
	//fmt.Printf("SourceChain %v\n", impTx.SourceChain.String())
	delTx := tx.UnsignedTx.(*platformvm.UnsignedAddDelegatorTx)

	fmt.Println(delTx.Ins[0].TxID.String())
	//replyStr, err := json.Marshal(bTx)
	//fmt.Println(replyStr)
	return nil
}

func testDecodingUtxo() error {
	byteStr := "0x000066a9593b87ddbadf7b5405fa8b584c5062295e8a000e3419e251acd5a6552f4d00000000dbcf890f77f49b96857648b72b77f9f82937f28a68704af05da0dc12ba53f2db00000007000000003b9ac99c000000000000000000000001000000013cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c815335e5"
	bytes, err := formatting.Decode(formatting.Hex, byteStr)
	if err != nil {
		return err
	}

	utxo := new(avax.UTXO)
	_, err = platformvm.Codec.Unmarshal(bytes, &utxo)

	if err != nil {
		return err
	}

	id := utxo.ID.String()
	fmt.Println(id)
	return nil
}

func testFlow() error {
	signers0, err := getSigners(testnetKey)
	if err != nil {
		return err
	}

	nodeID, err := ids.ShortFromPrefixedString("NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5", constants.NodeIDPrefix)

	if err != nil {
		return err
	}

	ctx, err := NewMyContext()
	if err != nil {
		return err
	}

	signers := [][]*avaCrypto.PrivateKeySECP256K1R{
		signers0,
	}
	expTxU := ctx.buildExportTx(0, 30, 500)
	expTxS := evm.Tx{UnsignedAtomicTx: &expTxU}
	expTxS.Sign(evm.Codec, signers)

	impTxU := ctx.buildImportTx(expTxS.UnsignedAtomicTx.(*evm.UnsignedExportTx))
	impTxS := platformvm.Tx{UnsignedTx: &impTxU}
	impTxS.Sign(platformvm.Codec, signers)

	delTxU := ctx.buildAddDelegatorTx(impTxS.UnsignedTx.(*platformvm.UnsignedImportTx), nodeID)
	delTxS := platformvm.Tx{UnsignedTx: &delTxU}
	delTxS.Sign(platformvm.Codec, signers)

	expTxID0 := expTxS.ID().String()
	fmt.Println(expTxID0)
	impTxID0 := impTxS.ID().String()
	fmt.Println(impTxID0)
	delTxID0 := delTxS.ID().String()
	fmt.Println(delTxID0)

	//cclient := evm.NewClient("http://localhost:9650", "C")
	pclient := platformvm.NewClient("http://localhost:9650")
	//expTxID, err := cclient.IssueTx(context.Background(), expTxS.Bytes())
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("ExportTx %v", expTxID)
	//time.Sleep(time.Second * 5)
	impTxID, err := pclient.IssueTx(context.Background(), impTxS.Bytes())
	if err != nil {
		return err
	}
	fmt.Printf("ImportTx %v", impTxID)
	time.Sleep(time.Second * 5)
	delTxID, err := pclient.IssueTx(context.Background(), delTxS.Bytes())
	if err != nil {
		return err
	}
	fmt.Printf("AddDelegatorTx %v", delTxID)

	return nil
}

func handler(c *cli.Context) error {
	//testExport()
	//testDecodingExport()
	//testStake()
	//testFilter()
	//err := startFilterer()
	//if err != nil {
	//	return err
	//}
	//sendTxn()
	//testDecodingPChain()
	//time.Sleep(35 * time.Second)
	//testDecodingUtxo()
	testFlow()

	return nil
}

func main() {
	app := &cli.App{
		Name:  PROG_NAME,
		Usage: "Handles the MPC operations needed for Avalanche",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     PARAM_URL,
				Required: true,
				Usage:    "The avalanche rpc url.",
			},
		},
		Action: handler,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
