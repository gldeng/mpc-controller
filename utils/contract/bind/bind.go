package bind

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"strings"
)

type ABI func() string

type Arg struct {
	Address    common.Address
	Caller     bind.ContractCaller
	Transactor bind.ContractTransactor
	Filterer   bind.ContractFilterer
	ABI        ABI
}

// Bind binds a generic wrapper to an already deployed contract.
func Bind(arg *Arg) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(arg.ABI()))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse ABI")
	}
	bound, err := bind.NewBoundContract(arg.Address, parsed, arg.Caller, arg.Transactor, arg.Filterer), nil
	return bound, errors.Wrapf(err, "failed to bound contract")
}
