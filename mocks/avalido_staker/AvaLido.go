// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package avalido_staker

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// AvaLidoMetaData contains all meta data concerning the AvaLido contract.
var AvaLidoMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mpcManagerAddress\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"_initiateStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManager\",\"outputs\":[{\"internalType\":\"contractMPCManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManagerAddress_\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// AvaLidoABI is the input ABI used to generate the binding from.
// Deprecated: Use AvaLidoMetaData.ABI instead.
var AvaLidoABI = AvaLidoMetaData.ABI

// AvaLido is an auto generated Go binding around an Ethereum contract.
type AvaLido struct {
	AvaLidoCaller     // Read-only binding to the contract
	AvaLidoTransactor // Write-only binding to the contract
	AvaLidoFilterer   // Log filterer for contract events
}

// AvaLidoCaller is an auto generated read-only Go binding around an Ethereum contract.
type AvaLidoCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AvaLidoTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AvaLidoTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AvaLidoFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AvaLidoFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AvaLidoSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AvaLidoSession struct {
	Contract     *AvaLido          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AvaLidoCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AvaLidoCallerSession struct {
	Contract *AvaLidoCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// AvaLidoTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AvaLidoTransactorSession struct {
	Contract     *AvaLidoTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// AvaLidoRaw is an auto generated low-level Go binding around an Ethereum contract.
type AvaLidoRaw struct {
	Contract *AvaLido // Generic contract binding to access the raw methods on
}

// AvaLidoCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AvaLidoCallerRaw struct {
	Contract *AvaLidoCaller // Generic read-only contract binding to access the raw methods on
}

// AvaLidoTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AvaLidoTransactorRaw struct {
	Contract *AvaLidoTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAvaLido creates a new instance of AvaLido, bound to a specific deployed contract.
func NewAvaLido(address common.Address, backend bind.ContractBackend) (*AvaLido, error) {
	contract, err := bindAvaLido(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AvaLido{AvaLidoCaller: AvaLidoCaller{contract: contract}, AvaLidoTransactor: AvaLidoTransactor{contract: contract}, AvaLidoFilterer: AvaLidoFilterer{contract: contract}}, nil
}

// NewAvaLidoCaller creates a new read-only instance of AvaLido, bound to a specific deployed contract.
func NewAvaLidoCaller(address common.Address, caller bind.ContractCaller) (*AvaLidoCaller, error) {
	contract, err := bindAvaLido(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AvaLidoCaller{contract: contract}, nil
}

// NewAvaLidoTransactor creates a new write-only instance of AvaLido, bound to a specific deployed contract.
func NewAvaLidoTransactor(address common.Address, transactor bind.ContractTransactor) (*AvaLidoTransactor, error) {
	contract, err := bindAvaLido(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AvaLidoTransactor{contract: contract}, nil
}

// NewAvaLidoFilterer creates a new log filterer instance of AvaLido, bound to a specific deployed contract.
func NewAvaLidoFilterer(address common.Address, filterer bind.ContractFilterer) (*AvaLidoFilterer, error) {
	contract, err := bindAvaLido(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AvaLidoFilterer{contract: contract}, nil
}

// bindAvaLido binds a generic wrapper to an already deployed contract.
func bindAvaLido(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AvaLidoABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AvaLido *AvaLidoRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AvaLido.Contract.AvaLidoCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AvaLido *AvaLidoRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvaLido.Contract.AvaLidoTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AvaLido *AvaLidoRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AvaLido.Contract.AvaLidoTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AvaLido *AvaLidoCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AvaLido.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AvaLido *AvaLidoTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvaLido.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AvaLido *AvaLidoTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AvaLido.Contract.contract.Transact(opts, method, params...)
}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_AvaLido *AvaLidoCaller) GetBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AvaLido.contract.Call(opts, &out, "getBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_AvaLido *AvaLidoSession) GetBalance() (*big.Int, error) {
	return _AvaLido.Contract.GetBalance(&_AvaLido.CallOpts)
}

// GetBalance is a free data retrieval call binding the contract method 0x12065fe0.
//
// Solidity: function getBalance() view returns(uint256)
func (_AvaLido *AvaLidoCallerSession) GetBalance() (*big.Int, error) {
	return _AvaLido.Contract.GetBalance(&_AvaLido.CallOpts)
}

// MpcManager is a free data retrieval call binding the contract method 0x846b3c8b.
//
// Solidity: function mpcManager() view returns(address)
func (_AvaLido *AvaLidoCaller) MpcManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AvaLido.contract.Call(opts, &out, "mpcManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MpcManager is a free data retrieval call binding the contract method 0x846b3c8b.
//
// Solidity: function mpcManager() view returns(address)
func (_AvaLido *AvaLidoSession) MpcManager() (common.Address, error) {
	return _AvaLido.Contract.MpcManager(&_AvaLido.CallOpts)
}

// MpcManager is a free data retrieval call binding the contract method 0x846b3c8b.
//
// Solidity: function mpcManager() view returns(address)
func (_AvaLido *AvaLidoCallerSession) MpcManager() (common.Address, error) {
	return _AvaLido.Contract.MpcManager(&_AvaLido.CallOpts)
}

// MpcManagerAddress is a free data retrieval call binding the contract method 0x7286bf2f.
//
// Solidity: function mpcManagerAddress_() view returns(address)
func (_AvaLido *AvaLidoCaller) MpcManagerAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AvaLido.contract.Call(opts, &out, "mpcManagerAddress_")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MpcManagerAddress is a free data retrieval call binding the contract method 0x7286bf2f.
//
// Solidity: function mpcManagerAddress_() view returns(address)
func (_AvaLido *AvaLidoSession) MpcManagerAddress() (common.Address, error) {
	return _AvaLido.Contract.MpcManagerAddress(&_AvaLido.CallOpts)
}

// MpcManagerAddress is a free data retrieval call binding the contract method 0x7286bf2f.
//
// Solidity: function mpcManagerAddress_() view returns(address)
func (_AvaLido *AvaLidoCallerSession) MpcManagerAddress() (common.Address, error) {
	return _AvaLido.Contract.MpcManagerAddress(&_AvaLido.CallOpts)
}

// InitiateStake is a paid mutator transaction binding the contract method 0xda52ab60.
//
// Solidity: function _initiateStake(uint256 amount) returns(uint256)
func (_AvaLido *AvaLidoTransactor) InitiateStake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _AvaLido.contract.Transact(opts, "_initiateStake", amount)
}

// InitiateStake is a paid mutator transaction binding the contract method 0xda52ab60.
//
// Solidity: function _initiateStake(uint256 amount) returns(uint256)
func (_AvaLido *AvaLidoSession) InitiateStake(amount *big.Int) (*types.Transaction, error) {
	return _AvaLido.Contract.InitiateStake(&_AvaLido.TransactOpts, amount)
}

// InitiateStake is a paid mutator transaction binding the contract method 0xda52ab60.
//
// Solidity: function _initiateStake(uint256 amount) returns(uint256)
func (_AvaLido *AvaLidoTransactorSession) InitiateStake(amount *big.Int) (*types.Transaction, error) {
	return _AvaLido.Contract.InitiateStake(&_AvaLido.TransactOpts, amount)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_AvaLido *AvaLidoTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvaLido.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_AvaLido *AvaLidoSession) Receive() (*types.Transaction, error) {
	return _AvaLido.Contract.Receive(&_AvaLido.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_AvaLido *AvaLidoTransactorSession) Receive() (*types.Transaction, error) {
	return _AvaLido.Contract.Receive(&_AvaLido.TransactOpts)
}
