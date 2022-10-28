package testingutils

import (
	"github.com/avalido/mpc-controller/contract"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

var (
	TestAddress = common.HexToAddress("0xa626f2e3a33b03459b84df1ac2756f2d9d44d0db")
)

func MakeEventRequestStarted(requestHash [32]byte, participantIndices *big.Int) *types.Log {
	abi, err := contract.MpcManagerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
	event := abi.Events["RequestStarted"]

	data, err := event.Inputs.Pack(requestHash, participantIndices)
	if err != nil {
		panic(err)
	}
	return &types.Log{
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
}
