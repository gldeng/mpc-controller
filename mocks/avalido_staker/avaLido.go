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
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mpcManagerAddress\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"_initiateStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManager\",\"outputs\":[{\"internalType\":\"contractIMPCManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManagerAddress_\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Sigs: map[string]string{
		"da52ab60": "_initiateStake(uint256)",
		"12065fe0": "getBalance()",
		"846b3c8b": "mpcManager()",
		"7286bf2f": "mpcManagerAddress_()",
	},
	Bin: "0x608060405260405161036038038061036083398101604081905261002291610051565b600080546001600160a01b039092166001600160a01b0319928316811790915560018054909216179055610081565b60006020828403121561006357600080fd5b81516001600160a01b038116811461007a57600080fd5b9392505050565b6102d0806100906000396000f3fe6080604052600436106100435760003560e01c806312065fe01461004f5780637286bf2f14610071578063846b3c8b146100a9578063da52ab60146100c957600080fd5b3661004a57005b600080fd5b34801561005b57600080fd5b50475b6040519081526020015b60405180910390f35b34801561007d57600080fd5b50600054610091906001600160a01b031681565b6040516001600160a01b039091168152602001610068565b3480156100b557600080fd5b50600154610091906001600160a01b031681565b3480156100d557600080fd5b5061005e6100e43660046101ca565b600080546040516001600160a01b039091169083156108fc0290849084818181858888f1935050505015801561011e573d6000803e3d6000fd5b50600061012c42601e6101e3565b9050600061013d62127500836101e3565b600154604080516060810190915260288082529293506001600160a01b0390911691636cfb1929919061027360208301398685856040518563ffffffff1660e01b81526004016101909493929190610209565b600060405180830381600087803b1580156101aa57600080fd5b505af11580156101be573d6000803e3d6000fd5b50959695505050505050565b6000602082840312156101dc57600080fd5b5035919050565b6000821982111561020457634e487b7160e01b600052601160045260246000fd5b500190565b608081526000855180608084015260005b8181101561023757602081890181015160a086840101520161021a565b8181111561024957600060a083860101525b506020830195909552506040810192909252606082015260a0601f909201601f1916010191905056fe4e6f646549442d50376f42324d636a42476757324e58585756596a56384a4544466f573978444535a2646970667358221220f1c61dccc890c43caa099468a464473d6b9b6250ef410dca170ca3192e608dc964736f6c634300080d0033",
}

// AvaLidoABI is the input ABI used to generate the binding from.
// Deprecated: Use AvaLidoMetaData.ABI instead.
var AvaLidoABI = AvaLidoMetaData.ABI

// Deprecated: Use AvaLidoMetaData.Sigs instead.
// AvaLidoFuncSigs maps the 4-byte function signature to its string representation.
var AvaLidoFuncSigs = AvaLidoMetaData.Sigs

// AvaLidoBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use AvaLidoMetaData.Bin instead.
var AvaLidoBin = AvaLidoMetaData.Bin

// DeployAvaLido deploys a new Ethereum contract, binding an instance of AvaLido to it.
func DeployAvaLido(auth *bind.TransactOpts, backend bind.ContractBackend, mpcManagerAddress common.Address) (common.Address, *types.Transaction, *AvaLido, error) {
	parsed, err := AvaLidoMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(AvaLidoBin), backend, mpcManagerAddress)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &AvaLido{AvaLidoCaller: AvaLidoCaller{contract: contract}, AvaLidoTransactor: AvaLidoTransactor{contract: contract}, AvaLidoFilterer: AvaLidoFilterer{contract: contract}}, nil
}

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

// IMPCManagerMetaData contains all meta data concerning the IMPCManager contract.
var IMPCManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"serveStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"6cfb1929": "serveStake(string,uint256,uint256,uint256)",
	},
}

// IMPCManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use IMPCManagerMetaData.ABI instead.
var IMPCManagerABI = IMPCManagerMetaData.ABI

// Deprecated: Use IMPCManagerMetaData.Sigs instead.
// IMPCManagerFuncSigs maps the 4-byte function signature to its string representation.
var IMPCManagerFuncSigs = IMPCManagerMetaData.Sigs

// IMPCManager is an auto generated Go binding around an Ethereum contract.
type IMPCManager struct {
	IMPCManagerCaller     // Read-only binding to the contract
	IMPCManagerTransactor // Write-only binding to the contract
	IMPCManagerFilterer   // Log filterer for contract events
}

// IMPCManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type IMPCManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMPCManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IMPCManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMPCManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IMPCManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMPCManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IMPCManagerSession struct {
	Contract     *IMPCManager      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IMPCManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IMPCManagerCallerSession struct {
	Contract *IMPCManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// IMPCManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IMPCManagerTransactorSession struct {
	Contract     *IMPCManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// IMPCManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type IMPCManagerRaw struct {
	Contract *IMPCManager // Generic contract binding to access the raw methods on
}

// IMPCManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IMPCManagerCallerRaw struct {
	Contract *IMPCManagerCaller // Generic read-only contract binding to access the raw methods on
}

// IMPCManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IMPCManagerTransactorRaw struct {
	Contract *IMPCManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIMPCManager creates a new instance of IMPCManager, bound to a specific deployed contract.
func NewIMPCManager(address common.Address, backend bind.ContractBackend) (*IMPCManager, error) {
	contract, err := bindIMPCManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IMPCManager{IMPCManagerCaller: IMPCManagerCaller{contract: contract}, IMPCManagerTransactor: IMPCManagerTransactor{contract: contract}, IMPCManagerFilterer: IMPCManagerFilterer{contract: contract}}, nil
}

// NewIMPCManagerCaller creates a new read-only instance of IMPCManager, bound to a specific deployed contract.
func NewIMPCManagerCaller(address common.Address, caller bind.ContractCaller) (*IMPCManagerCaller, error) {
	contract, err := bindIMPCManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IMPCManagerCaller{contract: contract}, nil
}

// NewIMPCManagerTransactor creates a new write-only instance of IMPCManager, bound to a specific deployed contract.
func NewIMPCManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*IMPCManagerTransactor, error) {
	contract, err := bindIMPCManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IMPCManagerTransactor{contract: contract}, nil
}

// NewIMPCManagerFilterer creates a new log filterer instance of IMPCManager, bound to a specific deployed contract.
func NewIMPCManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*IMPCManagerFilterer, error) {
	contract, err := bindIMPCManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IMPCManagerFilterer{contract: contract}, nil
}

// bindIMPCManager binds a generic wrapper to an already deployed contract.
func bindIMPCManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IMPCManagerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMPCManager *IMPCManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IMPCManager.Contract.IMPCManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMPCManager *IMPCManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMPCManager.Contract.IMPCManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMPCManager *IMPCManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMPCManager.Contract.IMPCManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMPCManager *IMPCManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IMPCManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMPCManager *IMPCManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMPCManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMPCManager *IMPCManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMPCManager.Contract.contract.Transact(opts, method, params...)
}

// ServeStake is a paid mutator transaction binding the contract method 0x6cfb1929.
//
// Solidity: function serveStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_IMPCManager *IMPCManagerTransactor) ServeStake(opts *bind.TransactOpts, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMPCManager.contract.Transact(opts, "serveStake", nodeID, amount, startTime, endTime)
}

// ServeStake is a paid mutator transaction binding the contract method 0x6cfb1929.
//
// Solidity: function serveStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_IMPCManager *IMPCManagerSession) ServeStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMPCManager.Contract.ServeStake(&_IMPCManager.TransactOpts, nodeID, amount, startTime, endTime)
}

// ServeStake is a paid mutator transaction binding the contract method 0x6cfb1929.
//
// Solidity: function serveStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_IMPCManager *IMPCManagerTransactorSession) ServeStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMPCManager.Contract.ServeStake(&_IMPCManager.TransactOpts, nodeID, amount, startTime, endTime)
}
