package contract

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ---------------------------------------------------------------------------------------------------------------------
// Interfaces regarding MpcManager

// Callers

type Caller interface {
	GetGroup(opts *bind.CallOpts, groupId [32]byte) ([][]byte, error)
	GetGroupIdByKey(opts *bind.CallOpts, publicKey []byte) ([32]byte, error)
	PrincipalTreasuryAddress(opts *bind.CallOpts) (common.Address, error)
	RewardTreasuryAddress(opts *bind.CallOpts) (common.Address, error)
}

// Transactor

type Transactor interface {
	JoinRequest(opts *bind.TransactOpts, participantId [32]byte, requestHash [32]byte) (*types.Transaction, error)
	ReportGeneratedKey(opts *bind.TransactOpts, participantId [32]byte, generatedPublicKey []byte) (*types.Transaction, error)
}
