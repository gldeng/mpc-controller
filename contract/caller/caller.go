package caller

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Caller interface {
	GetGroup(opts *bind.CallOpts, groupId [32]byte) ([][]byte, error)
	GetGroupIdByKey(opts *bind.CallOpts, publicKey []byte) ([32]byte, error)
	PrincipalTreasuryAddress(opts *bind.CallOpts) (common.Address, error)
	RewardTreasuryAddress(opts *bind.CallOpts) (common.Address, error)
}
