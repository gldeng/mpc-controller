// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package AvaLido

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

// IMpcManagerKeyInfo is an auto generated low-level Go binding around an user-defined struct.
type IMpcManagerKeyInfo struct {
	GroupId   [32]byte
	Confirmed bool
}

// AvaLidoMetaData contains all meta data concerning the AvaLido contract.
var AvaLidoMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"mpcManagerAddress\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"getBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"initiateStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManager\",\"outputs\":[{\"internalType\":\"contractIMpcManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"mpcManagerAddress_\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Sigs: map[string]string{
		"12065fe0": "getBalance()",
		"b80325d6": "initiateStake(uint256)",
		"846b3c8b": "mpcManager()",
		"7286bf2f": "mpcManagerAddress_()",
	},
	Bin: "0x608060405260405161036038038061036083398101604081905261002291610051565b600080546001600160a01b039092166001600160a01b0319928316811790915560018054909216179055610081565b60006020828403121561006357600080fd5b81516001600160a01b038116811461007a57600080fd5b9392505050565b6102d0806100906000396000f3fe6080604052600436106100435760003560e01c806312065fe01461004f5780637286bf2f14610071578063846b3c8b146100a9578063b80325d6146100c957600080fd5b3661004a57005b600080fd5b34801561005b57600080fd5b50475b6040519081526020015b60405180910390f35b34801561007d57600080fd5b50600054610091906001600160a01b031681565b6040516001600160a01b039091168152602001610068565b3480156100b557600080fd5b50600154610091906001600160a01b031681565b3480156100d557600080fd5b5061005e6100e43660046101ca565b600080546040516001600160a01b039091169083156108fc0290849084818181858888f1935050505015801561011e573d6000803e3d6000fd5b50600061012c42601e6101e3565b9050600061013d62127500836101e3565b600154604080516060810190915260288082529293506001600160a01b03909116916389060b34919061027360208301398685856040518563ffffffff1660e01b81526004016101909493929190610209565b600060405180830381600087803b1580156101aa57600080fd5b505af11580156101be573d6000803e3d6000fd5b50959695505050505050565b6000602082840312156101dc57600080fd5b5035919050565b6000821982111561020457634e487b7160e01b600052601160045260246000fd5b500190565b608081526000855180608084015260005b8181101561023757602081890181015160a086840101520161021a565b8181111561024957600060a083860101525b506020830195909552506040810192909252606082015260a0601f909201601f1916010191905056fe4e6f646549442d50376f42324d636a42476757324e58585756596a56384a4544466f573978444535a2646970667358221220b32b74b491bc0aaee6849db89a947bb2f01f03887b78d8020bc52c9d1049703064736f6c634300080e0033",
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

// InitiateStake is a paid mutator transaction binding the contract method 0xb80325d6.
//
// Solidity: function initiateStake(uint256 amount) returns(uint256)
func (_AvaLido *AvaLidoTransactor) InitiateStake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _AvaLido.contract.Transact(opts, "initiateStake", amount)
}

// InitiateStake is a paid mutator transaction binding the contract method 0xb80325d6.
//
// Solidity: function initiateStake(uint256 amount) returns(uint256)
func (_AvaLido *AvaLidoSession) InitiateStake(amount *big.Int) (*types.Transaction, error) {
	return _AvaLido.Contract.InitiateStake(&_AvaLido.TransactOpts, amount)
}

// InitiateStake is a paid mutator transaction binding the contract method 0xb80325d6.
//
// Solidity: function initiateStake(uint256 amount) returns(uint256)
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

// IMpcManagerMetaData contains all meta data concerning the IMpcManager contract.
var IMpcManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"publicKeys\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"createGroup\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"getGroup\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"participants\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"}],\"name\":\"getKey\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"confirmed\",\"type\":\"bool\"}],\"internalType\":\"structIMpcManager.KeyInfo\",\"name\":\"keyInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"requestKeygen\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"requestStake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"avaLidoAddress\",\"type\":\"address\"}],\"name\":\"setAvaLidoAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"dd6bd149": "createGroup(bytes[],uint256)",
		"b567d4ba": "getGroup(bytes32)",
		"7fed84f2": "getKey(bytes)",
		"e661d90d": "requestKeygen(bytes32)",
		"89060b34": "requestStake(string,uint256,uint256,uint256)",
		"78cdefae": "setAvaLidoAddress(address)",
	},
}

// IMpcManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use IMpcManagerMetaData.ABI instead.
var IMpcManagerABI = IMpcManagerMetaData.ABI

// Deprecated: Use IMpcManagerMetaData.Sigs instead.
// IMpcManagerFuncSigs maps the 4-byte function signature to its string representation.
var IMpcManagerFuncSigs = IMpcManagerMetaData.Sigs

// IMpcManager is an auto generated Go binding around an Ethereum contract.
type IMpcManager struct {
	IMpcManagerCaller     // Read-only binding to the contract
	IMpcManagerTransactor // Write-only binding to the contract
	IMpcManagerFilterer   // Log filterer for contract events
}

// IMpcManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type IMpcManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMpcManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IMpcManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMpcManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IMpcManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IMpcManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IMpcManagerSession struct {
	Contract     *IMpcManager      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IMpcManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IMpcManagerCallerSession struct {
	Contract *IMpcManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// IMpcManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IMpcManagerTransactorSession struct {
	Contract     *IMpcManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// IMpcManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type IMpcManagerRaw struct {
	Contract *IMpcManager // Generic contract binding to access the raw methods on
}

// IMpcManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IMpcManagerCallerRaw struct {
	Contract *IMpcManagerCaller // Generic read-only contract binding to access the raw methods on
}

// IMpcManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IMpcManagerTransactorRaw struct {
	Contract *IMpcManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIMpcManager creates a new instance of IMpcManager, bound to a specific deployed contract.
func NewIMpcManager(address common.Address, backend bind.ContractBackend) (*IMpcManager, error) {
	contract, err := bindIMpcManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IMpcManager{IMpcManagerCaller: IMpcManagerCaller{contract: contract}, IMpcManagerTransactor: IMpcManagerTransactor{contract: contract}, IMpcManagerFilterer: IMpcManagerFilterer{contract: contract}}, nil
}

// NewIMpcManagerCaller creates a new read-only instance of IMpcManager, bound to a specific deployed contract.
func NewIMpcManagerCaller(address common.Address, caller bind.ContractCaller) (*IMpcManagerCaller, error) {
	contract, err := bindIMpcManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IMpcManagerCaller{contract: contract}, nil
}

// NewIMpcManagerTransactor creates a new write-only instance of IMpcManager, bound to a specific deployed contract.
func NewIMpcManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*IMpcManagerTransactor, error) {
	contract, err := bindIMpcManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IMpcManagerTransactor{contract: contract}, nil
}

// NewIMpcManagerFilterer creates a new log filterer instance of IMpcManager, bound to a specific deployed contract.
func NewIMpcManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*IMpcManagerFilterer, error) {
	contract, err := bindIMpcManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IMpcManagerFilterer{contract: contract}, nil
}

// bindIMpcManager binds a generic wrapper to an already deployed contract.
func bindIMpcManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IMpcManagerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMpcManager *IMpcManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IMpcManager.Contract.IMpcManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMpcManager *IMpcManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMpcManager.Contract.IMpcManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMpcManager *IMpcManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMpcManager.Contract.IMpcManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IMpcManager *IMpcManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IMpcManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IMpcManager *IMpcManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IMpcManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IMpcManager *IMpcManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IMpcManager.Contract.contract.Transact(opts, method, params...)
}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_IMpcManager *IMpcManagerCaller) GetGroup(opts *bind.CallOpts, groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	var out []interface{}
	err := _IMpcManager.contract.Call(opts, &out, "getGroup", groupId)

	outstruct := new(struct {
		Participants [][]byte
		Threshold    *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Participants = *abi.ConvertType(out[0], new([][]byte)).(*[][]byte)
	outstruct.Threshold = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_IMpcManager *IMpcManagerSession) GetGroup(groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	return _IMpcManager.Contract.GetGroup(&_IMpcManager.CallOpts, groupId)
}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_IMpcManager *IMpcManagerCallerSession) GetGroup(groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	return _IMpcManager.Contract.GetGroup(&_IMpcManager.CallOpts, groupId)
}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_IMpcManager *IMpcManagerCaller) GetKey(opts *bind.CallOpts, publicKey []byte) (IMpcManagerKeyInfo, error) {
	var out []interface{}
	err := _IMpcManager.contract.Call(opts, &out, "getKey", publicKey)

	if err != nil {
		return *new(IMpcManagerKeyInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(IMpcManagerKeyInfo)).(*IMpcManagerKeyInfo)

	return out0, err

}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_IMpcManager *IMpcManagerSession) GetKey(publicKey []byte) (IMpcManagerKeyInfo, error) {
	return _IMpcManager.Contract.GetKey(&_IMpcManager.CallOpts, publicKey)
}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_IMpcManager *IMpcManagerCallerSession) GetKey(publicKey []byte) (IMpcManagerKeyInfo, error) {
	return _IMpcManager.Contract.GetKey(&_IMpcManager.CallOpts, publicKey)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_IMpcManager *IMpcManagerTransactor) CreateGroup(opts *bind.TransactOpts, publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _IMpcManager.contract.Transact(opts, "createGroup", publicKeys, threshold)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_IMpcManager *IMpcManagerSession) CreateGroup(publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _IMpcManager.Contract.CreateGroup(&_IMpcManager.TransactOpts, publicKeys, threshold)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_IMpcManager *IMpcManagerTransactorSession) CreateGroup(publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _IMpcManager.Contract.CreateGroup(&_IMpcManager.TransactOpts, publicKeys, threshold)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_IMpcManager *IMpcManagerTransactor) RequestKeygen(opts *bind.TransactOpts, groupId [32]byte) (*types.Transaction, error) {
	return _IMpcManager.contract.Transact(opts, "requestKeygen", groupId)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_IMpcManager *IMpcManagerSession) RequestKeygen(groupId [32]byte) (*types.Transaction, error) {
	return _IMpcManager.Contract.RequestKeygen(&_IMpcManager.TransactOpts, groupId)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_IMpcManager *IMpcManagerTransactorSession) RequestKeygen(groupId [32]byte) (*types.Transaction, error) {
	return _IMpcManager.Contract.RequestKeygen(&_IMpcManager.TransactOpts, groupId)
}

// RequestStake is a paid mutator transaction binding the contract method 0x89060b34.
//
// Solidity: function requestStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) payable returns()
func (_IMpcManager *IMpcManagerTransactor) RequestStake(opts *bind.TransactOpts, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMpcManager.contract.Transact(opts, "requestStake", nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x89060b34.
//
// Solidity: function requestStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) payable returns()
func (_IMpcManager *IMpcManagerSession) RequestStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMpcManager.Contract.RequestStake(&_IMpcManager.TransactOpts, nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x89060b34.
//
// Solidity: function requestStake(string nodeID, uint256 amount, uint256 startTime, uint256 endTime) payable returns()
func (_IMpcManager *IMpcManagerTransactorSession) RequestStake(nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _IMpcManager.Contract.RequestStake(&_IMpcManager.TransactOpts, nodeID, amount, startTime, endTime)
}

// SetAvaLidoAddress is a paid mutator transaction binding the contract method 0x78cdefae.
//
// Solidity: function setAvaLidoAddress(address avaLidoAddress) returns()
func (_IMpcManager *IMpcManagerTransactor) SetAvaLidoAddress(opts *bind.TransactOpts, avaLidoAddress common.Address) (*types.Transaction, error) {
	return _IMpcManager.contract.Transact(opts, "setAvaLidoAddress", avaLidoAddress)
}

// SetAvaLidoAddress is a paid mutator transaction binding the contract method 0x78cdefae.
//
// Solidity: function setAvaLidoAddress(address avaLidoAddress) returns()
func (_IMpcManager *IMpcManagerSession) SetAvaLidoAddress(avaLidoAddress common.Address) (*types.Transaction, error) {
	return _IMpcManager.Contract.SetAvaLidoAddress(&_IMpcManager.TransactOpts, avaLidoAddress)
}

// SetAvaLidoAddress is a paid mutator transaction binding the contract method 0x78cdefae.
//
// Solidity: function setAvaLidoAddress(address avaLidoAddress) returns()
func (_IMpcManager *IMpcManagerTransactorSession) SetAvaLidoAddress(avaLidoAddress common.Address) (*types.Transaction, error) {
	return _IMpcManager.Contract.SetAvaLidoAddress(&_IMpcManager.TransactOpts, avaLidoAddress)
}
