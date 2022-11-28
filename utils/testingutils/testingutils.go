package testingutils

import (
	"github.com/avalido/mpc-controller/contract"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

var (
	TestAddress = common.HexToAddress("0xa626f2e3a33b03459b84df1ac2756f2d9d44d0db")
)

func MakeEventParticipantAdded(pubKey []byte, groupId [32]byte, index *big.Int) *contract.MpcManagerParticipantAdded {
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	event := abi.Events["ParticipantAdded"]

	data, err := event.Inputs.NonIndexed().Pack(groupId, index)
	if err != nil {
		panic(err)
	}
	log := types.Log{
		Address: common.HexToAddress("0xa626f2e3a33b03459b84df1ac2756f2d9d44d0db"),
		Topics: []common.Hash{
			event.ID,
			common.BytesToHash(crypto.Keccak256(pubKey)),
		},
		Data:        data,
		BlockNumber: 0x3802,
		TxHash:      common.HexToHash("0xc8ddd3b3a163ede531ef5f9762825358d909b3f5328d4586a20f724d9cf1e661"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x31134b0cb11e04161f6b59f1406f7b561252566c38252073c4812c0197f050ae"),
		Index:       0,
		Removed:     false,
	}
	e := &contract.MpcManagerParticipantAdded{}
	e.PublicKey = log.Topics[1]
	err = abi.UnpackIntoInterface(e, "ParticipantAdded", data)
	if err != nil {
		panic(err)
	}
	e.Raw = log
	return e
}

func MakeEventKeygenRequestAdded(groupId [32]byte, reqNo *big.Int) *contract.MpcManagerKeygenRequestAdded {
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	event := abi.Events["KeygenRequestAdded"]

	data, err := event.Inputs.NonIndexed().Pack(reqNo)
	if err != nil {
		panic(err)
	}
	log := types.Log{
		Address: common.HexToAddress("0xa626f2e3a33b03459b84df1ac2756f2d9d44d0db"),
		Topics: []common.Hash{
			event.ID,
			common.BytesToHash(crypto.Keccak256(groupId[:])),
		},
		Data:        data,
		BlockNumber: 0x3802,
		TxHash:      common.HexToHash("0xc8ddd3b3a163ede531ef5f9762825358d909b3f5328d4586a20f724d9cf1e661"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x31134b0cb11e04161f6b59f1406f7b561252566c38252073c4812c0197f050ae"),
		Index:       0,
		Removed:     false,
	}
	e := &contract.MpcManagerKeygenRequestAdded{}
	err = abi.UnpackIntoInterface(e, "KeygenRequestAdded", data)
	if err != nil {
		panic(err)
	}
	e.Raw = log
	return e
}

func MakeEventStakeRequestAdded(reqNo int64, pubKey []byte) *contract.MpcManagerStakeRequestAdded {
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	event := abi.Events["StakeRequestAdded"]

	amount := new(big.Int)
	amount.SetString("999000000000", 10)
	startTime := new(big.Int)
	startTime.SetInt64(1663315662)
	endTime := new(big.Int)
	endTime.SetInt64(1694830062)
	data, err := event.Inputs.NonIndexed().Pack(
		big.NewInt(reqNo), "NodeID-P7oB2McjBGgW2NXXWVYjV8JEDFoW9xDE5",
		amount,
		startTime,
		endTime,
	)

	if err != nil {
		panic(err)
	}
	log := types.Log{
		Address: common.HexToAddress("0xa626f2e3a33b03459b84df1ac2756f2d9d44d0db"),
		Topics: []common.Hash{
			event.ID,
			common.BytesToHash(crypto.Keccak256(pubKey)),
		},
		Data:        data,
		BlockNumber: 0x3802,
		TxHash:      common.HexToHash("0xc8ddd3b3a163ede531ef5f9762825358d909b3f5328d4586a20f724d9cf1e661"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x31134b0cb11e04161f6b59f1406f7b561252566c38252073c4812c0197f050ae"),
		Index:       0,
		Removed:     false,
	}
	e := &contract.MpcManagerStakeRequestAdded{}
	err = abi.UnpackIntoInterface(e, "StakeRequestAdded", data)
	if err != nil {
		panic(err)
	}
	e.Raw = log
	return e
}

func MakeEventRequestStarted(requestHash [32]byte, participantIndices *big.Int) *contract.MpcManagerRequestStarted {
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	event := abi.Events["RequestStarted"]

	data, err := event.Inputs.NonIndexed().Pack(requestHash, participantIndices)
	if err != nil {
		panic(err)
	}

	log := types.Log{
		Address:     common.HexToAddress("0xa626f2e3a33b03459b84df1ac2756f2d9d44d0db"),
		Topics:      []common.Hash{event.ID},
		Data:        data,
		BlockNumber: 0x3802,
		TxHash:      common.HexToHash("0xc8ddd3b3a163ede531ef5f9762825358d909b3f5328d4586a20f724d9cf1e661"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x31134b0cb11e04161f6b59f1406f7b561252566c38252073c4812c0197f050ae"),
		Index:       0,
		Removed:     false,
	}
	e := &contract.MpcManagerRequestStarted{}
	err = abi.UnpackIntoInterface(e, "RequestStarted", data)
	if err != nil {
		panic(err)
	}
	e.Raw = log
	return e
}

func MakeEventKeyGenerated(groupId [32]byte, publicKey []byte) *contract.MpcManagerKeyGenerated {
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	event := abi.Events["KeyGenerated"]

	data, err := event.Inputs.NonIndexed().Pack(publicKey)
	if err != nil {
		panic(err)
	}

	log := types.Log{
		Address: common.HexToAddress("0xa626f2e3a33b03459b84df1ac2756f2d9d44d0db"),
		Topics: []common.Hash{
			event.ID,
			common.BytesToHash(crypto.Keccak256(groupId[:])),
		},
		Data:        data,
		BlockNumber: 0x3802,
		TxHash:      common.HexToHash("0xc8ddd3b3a163ede531ef5f9762825358d909b3f5328d4586a20f724d9cf1e661"),
		TxIndex:     0,
		BlockHash:   common.HexToHash("0x31134b0cb11e04161f6b59f1406f7b561252566c38252073c4812c0197f050ae"),
		Index:       0,
		Removed:     false,
	}
	e := &contract.MpcManagerKeyGenerated{}
	err = abi.UnpackIntoInterface(e, "KeyGenerated", data)
	if err != nil {
		panic(err)
	}
	e.Raw = log
	return e
}
