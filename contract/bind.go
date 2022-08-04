package contract

import (
	myBind "github.com/avalido/mpc-controller/utils/contract/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var ABI = MpcManagerMetaData.ABI

func BindMpcManager(address common.Address, Caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	arg := myBind.Arg{
		Address:    address,
		Caller:     Caller,
		Transactor: transactor,
		Filterer:   filterer,
		ABI:        func() string { return ABI },
	}
	bound, err := myBind.Bind(&arg)
	return bound, errors.Wrapf(err, "failed to bind MpcManager")
}

func BindMpcManagerCaller(address common.Address, Caller bind.ContractCaller) (*bind.BoundContract, error) {
	arg := myBind.Arg{
		Address:    address,
		Caller:     Caller,
		Transactor: nil,
		Filterer:   nil,
		ABI:        func() string { return ABI },
	}
	bound, err := myBind.Bind(&arg)
	return bound, errors.Wrapf(err, "failed to bind MpcManager")
}

func BindMpcManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
	arg := myBind.Arg{
		Address:    address,
		Caller:     nil,
		Transactor: transactor,
		Filterer:   nil,
		ABI:        func() string { return ABI },
	}
	bound, err := myBind.Bind(&arg)
	return bound, errors.Wrapf(err, "failed to bind MpcManager")
}

func BindMpcManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	arg := myBind.Arg{
		Address:    address,
		Caller:     nil,
		Transactor: nil,
		Filterer:   filterer,
		ABI:        func() string { return ABI },
	}
	bound, err := myBind.Bind(&arg)
	return bound, errors.Wrapf(err, "failed to bind MpcManager")
}
