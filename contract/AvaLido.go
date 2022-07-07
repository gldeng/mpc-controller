// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mpcManagerAddress\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initiateStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManager\",\"outputs\":[{\"internalType\":\"contractIMpcManagerSimple\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManagerAddress_\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Sigs: map[string]string{
		"12065fe0": "getBalance()",
		"db4567d3": "initiateStake()",
		"846b3c8b": "mpcManager()",
		"7286bf2f": "mpcManagerAddress_()",
	},
	Bin: "0x608060405260405161031f38038061031f83398101604081905261002291610051565b600080546001600160a01b039092166001600160a01b0319928316811790915560018054909216179055610081565b60006020828403121561006357600080fd5b81516001600160a01b038116811461007a57600080fd5b9392505050565b61028f806100906000396000f3fe6080604052600436106100435760003560e01c806312065fe01461004f5780637286bf2f14610071578063846b3c8b146100a9578063db4567d3146100c957600080fd5b3661004a57005b600080fd5b34801561005b57600080fd5b50475b6040519081526020015b60405180910390f35b34801561007d57600080fd5b50600054610091906001600160a01b031681565b6040516001600160a01b039091168152602001610068565b3480156100b557600080fd5b50600154610091906001600160a01b031681565b3480156100d557600080fd5b5061005e6000806100e84261012c6101a2565b905060006100f962127500836101a2565b600154604080516060810190915260288082529293506001600160a01b03909116916389060b349168015af1d78b58c4000091610232602083013968015af1d78b58c4000086866040518663ffffffff1660e01b815260040161015f94939291906101c8565b6000604051808303818588803b15801561017857600080fd5b505af115801561018c573d6000803e3d6000fd5b505050505068015af1d78b58c400009250505090565b600082198211156101c357634e487b7160e01b600052601160045260246000fd5b500190565b608081526000855180608084015260005b818110156101f657602081890181015160a08684010152016101d9565b8181111561020857600060a083860101525b506020830195909552506040810192909252606082015260a0601f909201601f1916010191905056fe4e6f646549442d50376f42324d636a42476757324e58585756596a56384a4544466f573978444535a2646970667358221220158b3fa69ece76eb7a8e05e6b0aa7cedd59e27ae46af5025eefe48ab0ad0b95e64736f6c634300080a0033",
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

// InitiateStake is a paid mutator transaction binding the contract method 0xdb4567d3.
//
// Solidity: function initiateStake() returns(uint256)
func (_AvaLido *AvaLidoTransactor) InitiateStake(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AvaLido.contract.Transact(opts, "initiateStake")
}

// InitiateStake is a paid mutator transaction binding the contract method 0xdb4567d3.
//
// Solidity: function initiateStake() returns(uint256)
func (_AvaLido *AvaLidoSession) InitiateStake() (*types.Transaction, error) {
	return _AvaLido.Contract.InitiateStake(&_AvaLido.TransactOpts)
}

// InitiateStake is a paid mutator transaction binding the contract method 0xdb4567d3.
//
// Solidity: function initiateStake() returns(uint256)
func (_AvaLido *AvaLidoTransactorSession) InitiateStake() (*types.Transaction, error) {
	return _AvaLido.Contract.InitiateStake(&_AvaLido.TransactOpts)
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

// IMpcManagerSimpleMetaData contains all meta data concerning the IMpcManagerSimple contract.
var IMpcManagerSimpleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"requestStake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"89060b34": "requestStake(string,uint256,uint256,uint256)",
	},
}

// IMpcManagerSimpleABI is the input ABI used to generate the binding from.
// Deprecated: Use IMpcManagerSimpleMetaData.ABI instead.
var IMpcManagerSimpleABI = IMpcManagerSimpleMetaData.ABI

// Deprecated: Use IMpcManagerSimpleMetaData.Sigs instead.
// IMpcManagerSimpleFuncSigs maps the 4-byte function signature to its string representation.
var IMpcManagerSimpleFuncSigs = IMpcManagerSimpleMetaData.Sigs

// IMpcManagerSimple is an auto generated Go binding around an Ethereum contract.
type IMpcManagerSimple struct {
	IMpcManagerSimpleCaller     // Read-only binding to the contract
	IMpcManagerSimpleTransactor // Write-only binding to the contract
	IMpcManagerSimpleFilterer   // Log filterer for contract events
}

// IMpcManagerSimpleCaller is an auto generated read-only Go binding around an Ethereum contract.
type IMpcManagerSimpleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMpcManagerSimpleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IMpcManagerSimpleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMpcManagerSimpleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IMpcManagerSimpleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMpcManagerSimpleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IMpcManagerSimpleSession struct {
	Contract     *IMpcManagerSimple // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// IMpcManagerSimpleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IMpcManagerSimpleCallerSession struct {
	Contract *IMpcManagerSimpleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// IMpcManagerSimpleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IMpcManagerSimpleTransactorSession struct {
	Contract     *IMpcManagerSimpleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// IMpcManagerSimpleRaw is an auto generated low-level Go binding around an Ethereum contract.
type IMpcManagerSimpleRaw struct {
	Contract *IMpcManagerSimple // Generic contract binding to access the raw methods on
}

// IMpcManagerSimpleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IMpcManagerSimpleCallerRaw struct {
	Contract *IMpcManagerSimpleCaller // Generic read-only contract binding to access the raw methods on
}

// IMpcManagerSimpleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IMpcManagerSimpleTransactorRaw struct {
	Contract *IMpcManagerSimpleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIMpcManagerSimple creates a new instance of IMpcManagerSimple, bound to a specific deployed contract.
func NewIMpcManagerSimple(address common.Address, backend bind.ContractBackend) (*IMpcManagerSimple, error) {
	contract, err := bindIMpcManagerSimple(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IMpcManagerSimple{IMpcManagerSimpleCaller: IMpcManagerSimpleCaller{contract: contract}, IMpcManagerSimpleTransactor: IMpcManagerSimpleTransactor{contract: contract}, IMpcManagerSimpleFilterer: IMpcManagerSimpleFilterer{contract: contract}}, nil
}

// NewIMpcManagerSimpleCaller creates a new read-only instance of IMpcManagerSimple, bound to a specific deployed contract.
func NewIMpcManagerSimpleCaller(address common.Address, caller bind.ContractCaller) (*IMpcManagerSimpleCaller, error) {
	contract, err := bindIMpcManagerSimple(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IMpcManagerSimpleCaller{contract: contract}, nil
}

// NewIMpcManagerSimpleTransactor creates a new write-only instance of IMpcManagerSimple, bound to a specific deployed contract.
func NewIMpcManagerSimpleTransactor(address common.Address, transactor bind.ContractTransactor) (*IMpcManagerSimpleTransactor, error) {
	contract, err := bindIMpcManagerSimple(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IMpcManagerSimpleTransactor{contract: contract}, nil
}

// NewIMpcManagerSimpleFilterer creates a new log filterer instance of IMpcManagerSimple, bound to a specific deployed contract.
func NewIMpcManagerSimpleFilterer(address common.Address, filterer bind.ContractFilterer) (*IMpcManagerSimpleFilterer, error) {
	contract, err := bindIMpcManagerSimple(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IMpcManagerSimpleFilterer{contract: contract}, nil
}

// bindIMpcManagerSimple binds a generic wrapper to an already deployed contract.
func bindIMpcManagerSimple(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IMpcManagerSimpleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMpcManagerSimple *IMpcManagerSimpleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IMpcManagerSimple.Contract.IMpcManagerSimpleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMpcManagerSimple *IMpcManagerSimpleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMpcManagerSimple.Contract.IMpcManagerSimpleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMpcManagerSimple *IMpcManagerSimpleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMpcManagerSimple.Contract.IMpcManagerSimpleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMpcManagerSimple *IMpcManagerSimpleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IMpcManagerSimple.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMpcManagerSimple *IMpcManagerSimpleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMpcManagerSimple.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMpcManagerSimple *IMpcManagerSimpleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMpcManagerSimple.Contract.contract.Transact(opts, method, params...)
}

// RequestStake is a paid mutator transaction binding the contract method 0x89060b34.
//
// Solidity: function requestStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) payable returns()
func (_IMpcManagerSimple *IMpcManagerSimpleTransactor) RequestStake(opts *bind.TransactOpts, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMpcManagerSimple.contract.Transact(opts, "requestStake", nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x89060b34.
//
// Solidity: function requestStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) payable returns()
func (_IMpcManagerSimple *IMpcManagerSimpleSession) RequestStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMpcManagerSimple.Contract.RequestStake(&_IMpcManagerSimple.TransactOpts, nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x89060b34.
//
// Solidity: function requestStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) payable returns()
func (_IMpcManagerSimple *IMpcManagerSimpleTransactorSession) RequestStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMpcManagerSimple.Contract.RequestStake(&_IMpcManagerSimple.TransactOpts, nodeID, amount, startTime, endTime)
}
