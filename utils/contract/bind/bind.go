package bind

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"strings"
)

type BoundContract interface {
	BoundCaller
	BoundTransactor
	BoundFilterer
}

type BoundCaller interface {
	Call(opts *bind.CallOpts, results *[]interface{}, method string, params ...interface{}) error
}

type BoundTransactor interface {
	Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error)
	RawTransact(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error)
	Transfer(opts *bind.TransactOpts) (*types.Transaction, error)
}

type BoundFilterer interface {
	FilterLogs(opts *bind.FilterOpts, name string, query ...[]interface{}) (chan types.Log, event.Subscription, error)
	WatchLogs(opts *bind.WatchOpts, name string, query ...[]interface{}) (chan types.Log, event.Subscription, error)
	UnpackLog(out interface{}, event string, log types.Log) error
	UnpackLogIntoMap(out map[string]interface{}, event string, log types.Log) error
}

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
