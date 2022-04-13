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

// MpcCoordinatorKeyInfo is an auto generated low-level Go binding around an user-defined struct.
type MpcCoordinatorKeyInfo struct {
	GroupId   [32]byte
	Confirmed bool
}

// MpcCoordinatorMetaData contains all meta data concerning the MpcCoordinator contract.
var MpcCoordinatorMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"}],\"name\":\"KeyGenerated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"KeygenRequestAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"ParticipantAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"SignRequestAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"SignRequestStarted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"StakeRequestAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"participantIndices\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"StakeRequestStarted\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes[]\",\"name\":\"publicKeys\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"createGroup\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"getGroup\",\"outputs\":[{\"internalType\":\"bytes[]\",\"name\":\"participants\",\"type\":\"bytes[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"}],\"name\":\"getKey\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"confirmed\",\"type\":\"bool\"}],\"internalType\":\"structMpcCoordinator.KeyInfo\",\"name\":\"keyInfo\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requestId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"myIndex\",\"type\":\"uint256\"}],\"name\":\"joinRequest\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"myIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"generatedPublicKey\",\"type\":\"bytes\"}],\"name\":\"reportGeneratedKey\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"groupId\",\"type\":\"bytes32\"}],\"name\":\"requestKeygen\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"requestSign\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"publicKey\",\"type\":\"bytes\"},{\"internalType\":\"string\",\"name\":\"nodeID\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"}],\"name\":\"requestStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// MpcCoordinatorABI is the input ABI used to generate the binding from.
// Deprecated: Use MpcCoordinatorMetaData.ABI instead.
var MpcCoordinatorABI = MpcCoordinatorMetaData.ABI

// MpcCoordinator is an auto generated Go binding around an Ethereum contract.
type MpcCoordinator struct {
	MpcCoordinatorCaller     // Read-only binding to the contract
	MpcCoordinatorTransactor // Write-only binding to the contract
	MpcCoordinatorFilterer   // Log filterer for contract events
}

// MpcCoordinatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type MpcCoordinatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MpcCoordinatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MpcCoordinatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MpcCoordinatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MpcCoordinatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MpcCoordinatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MpcCoordinatorSession struct {
	Contract     *MpcCoordinator   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MpcCoordinatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MpcCoordinatorCallerSession struct {
	Contract *MpcCoordinatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// MpcCoordinatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MpcCoordinatorTransactorSession struct {
	Contract     *MpcCoordinatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// MpcCoordinatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type MpcCoordinatorRaw struct {
	Contract *MpcCoordinator // Generic contract binding to access the raw methods on
}

// MpcCoordinatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MpcCoordinatorCallerRaw struct {
	Contract *MpcCoordinatorCaller // Generic read-only contract binding to access the raw methods on
}

// MpcCoordinatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MpcCoordinatorTransactorRaw struct {
	Contract *MpcCoordinatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMpcCoordinator creates a new instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinator(address common.Address, backend bind.ContractBackend) (*MpcCoordinator, error) {
	contract, err := bindMpcCoordinator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinator{MpcCoordinatorCaller: MpcCoordinatorCaller{contract: contract}, MpcCoordinatorTransactor: MpcCoordinatorTransactor{contract: contract}, MpcCoordinatorFilterer: MpcCoordinatorFilterer{contract: contract}}, nil
}

// NewMpcCoordinatorCaller creates a new read-only instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinatorCaller(address common.Address, caller bind.ContractCaller) (*MpcCoordinatorCaller, error) {
	contract, err := bindMpcCoordinator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorCaller{contract: contract}, nil
}

// NewMpcCoordinatorTransactor creates a new write-only instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinatorTransactor(address common.Address, transactor bind.ContractTransactor) (*MpcCoordinatorTransactor, error) {
	contract, err := bindMpcCoordinator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorTransactor{contract: contract}, nil
}

// NewMpcCoordinatorFilterer creates a new log filterer instance of MpcCoordinator, bound to a specific deployed contract.
func NewMpcCoordinatorFilterer(address common.Address, filterer bind.ContractFilterer) (*MpcCoordinatorFilterer, error) {
	contract, err := bindMpcCoordinator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorFilterer{contract: contract}, nil
}

// bindMpcCoordinator binds a generic wrapper to an already deployed contract.
func bindMpcCoordinator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MpcCoordinatorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MpcCoordinator *MpcCoordinatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MpcCoordinator.Contract.MpcCoordinatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MpcCoordinator *MpcCoordinatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.MpcCoordinatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MpcCoordinator *MpcCoordinatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.MpcCoordinatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MpcCoordinator *MpcCoordinatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MpcCoordinator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MpcCoordinator *MpcCoordinatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MpcCoordinator *MpcCoordinatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.contract.Transact(opts, method, params...)
}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_MpcCoordinator *MpcCoordinatorCaller) GetGroup(opts *bind.CallOpts, groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getGroup", groupId)

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
func (_MpcCoordinator *MpcCoordinatorSession) GetGroup(groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	return _MpcCoordinator.Contract.GetGroup(&_MpcCoordinator.CallOpts, groupId)
}

// GetGroup is a free data retrieval call binding the contract method 0xb567d4ba.
//
// Solidity: function getGroup(bytes32 groupId) view returns(bytes[] participants, uint256 threshold)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetGroup(groupId [32]byte) (struct {
	Participants [][]byte
	Threshold    *big.Int
}, error) {
	return _MpcCoordinator.Contract.GetGroup(&_MpcCoordinator.CallOpts, groupId)
}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_MpcCoordinator *MpcCoordinatorCaller) GetKey(opts *bind.CallOpts, publicKey []byte) (MpcCoordinatorKeyInfo, error) {
	var out []interface{}
	err := _MpcCoordinator.contract.Call(opts, &out, "getKey", publicKey)

	if err != nil {
		return *new(MpcCoordinatorKeyInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(MpcCoordinatorKeyInfo)).(*MpcCoordinatorKeyInfo)

	return out0, err

}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_MpcCoordinator *MpcCoordinatorSession) GetKey(publicKey []byte) (MpcCoordinatorKeyInfo, error) {
	return _MpcCoordinator.Contract.GetKey(&_MpcCoordinator.CallOpts, publicKey)
}

// GetKey is a free data retrieval call binding the contract method 0x7fed84f2.
//
// Solidity: function getKey(bytes publicKey) view returns((bytes32,bool) keyInfo)
func (_MpcCoordinator *MpcCoordinatorCallerSession) GetKey(publicKey []byte) (MpcCoordinatorKeyInfo, error) {
	return _MpcCoordinator.Contract.GetKey(&_MpcCoordinator.CallOpts, publicKey)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) CreateGroup(opts *bind.TransactOpts, publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "createGroup", publicKeys, threshold)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_MpcCoordinator *MpcCoordinatorSession) CreateGroup(publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.CreateGroup(&_MpcCoordinator.TransactOpts, publicKeys, threshold)
}

// CreateGroup is a paid mutator transaction binding the contract method 0xdd6bd149.
//
// Solidity: function createGroup(bytes[] publicKeys, uint256 threshold) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) CreateGroup(publicKeys [][]byte, threshold *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.CreateGroup(&_MpcCoordinator.TransactOpts, publicKeys, threshold)
}

// JoinRequest is a paid mutator transaction binding the contract method 0x2ed92550.
//
// Solidity: function joinRequest(uint256 requestId, uint256 myIndex) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) JoinRequest(opts *bind.TransactOpts, requestId *big.Int, myIndex *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "joinRequest", requestId, myIndex)
}

// JoinRequest is a paid mutator transaction binding the contract method 0x2ed92550.
//
// Solidity: function joinRequest(uint256 requestId, uint256 myIndex) returns()
func (_MpcCoordinator *MpcCoordinatorSession) JoinRequest(requestId *big.Int, myIndex *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.JoinRequest(&_MpcCoordinator.TransactOpts, requestId, myIndex)
}

// JoinRequest is a paid mutator transaction binding the contract method 0x2ed92550.
//
// Solidity: function joinRequest(uint256 requestId, uint256 myIndex) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) JoinRequest(requestId *big.Int, myIndex *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.JoinRequest(&_MpcCoordinator.TransactOpts, requestId, myIndex)
}

// ReportGeneratedKey is a paid mutator transaction binding the contract method 0xfae3a93c.
//
// Solidity: function reportGeneratedKey(bytes32 groupId, uint256 myIndex, bytes generatedPublicKey) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) ReportGeneratedKey(opts *bind.TransactOpts, groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "reportGeneratedKey", groupId, myIndex, generatedPublicKey)
}

// ReportGeneratedKey is a paid mutator transaction binding the contract method 0xfae3a93c.
//
// Solidity: function reportGeneratedKey(bytes32 groupId, uint256 myIndex, bytes generatedPublicKey) returns()
func (_MpcCoordinator *MpcCoordinatorSession) ReportGeneratedKey(groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.ReportGeneratedKey(&_MpcCoordinator.TransactOpts, groupId, myIndex, generatedPublicKey)
}

// ReportGeneratedKey is a paid mutator transaction binding the contract method 0xfae3a93c.
//
// Solidity: function reportGeneratedKey(bytes32 groupId, uint256 myIndex, bytes generatedPublicKey) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) ReportGeneratedKey(groupId [32]byte, myIndex *big.Int, generatedPublicKey []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.ReportGeneratedKey(&_MpcCoordinator.TransactOpts, groupId, myIndex, generatedPublicKey)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) RequestKeygen(opts *bind.TransactOpts, groupId [32]byte) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "requestKeygen", groupId)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_MpcCoordinator *MpcCoordinatorSession) RequestKeygen(groupId [32]byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestKeygen(&_MpcCoordinator.TransactOpts, groupId)
}

// RequestKeygen is a paid mutator transaction binding the contract method 0xe661d90d.
//
// Solidity: function requestKeygen(bytes32 groupId) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) RequestKeygen(groupId [32]byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestKeygen(&_MpcCoordinator.TransactOpts, groupId)
}

// RequestSign is a paid mutator transaction binding the contract method 0x2f7e3d17.
//
// Solidity: function requestSign(bytes publicKey, bytes message) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) RequestSign(opts *bind.TransactOpts, publicKey []byte, message []byte) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "requestSign", publicKey, message)
}

// RequestSign is a paid mutator transaction binding the contract method 0x2f7e3d17.
//
// Solidity: function requestSign(bytes publicKey, bytes message) returns()
func (_MpcCoordinator *MpcCoordinatorSession) RequestSign(publicKey []byte, message []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestSign(&_MpcCoordinator.TransactOpts, publicKey, message)
}

// RequestSign is a paid mutator transaction binding the contract method 0x2f7e3d17.
//
// Solidity: function requestSign(bytes publicKey, bytes message) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) RequestSign(publicKey []byte, message []byte) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestSign(&_MpcCoordinator.TransactOpts, publicKey, message)
}

// RequestStake is a paid mutator transaction binding the contract method 0x2dbf0344.
//
// Solidity: function requestStake(bytes publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorTransactor) RequestStake(opts *bind.TransactOpts, publicKey []byte, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.contract.Transact(opts, "requestStake", publicKey, nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x2dbf0344.
//
// Solidity: function requestStake(bytes publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorSession) RequestStake(publicKey []byte, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestStake(&_MpcCoordinator.TransactOpts, publicKey, nodeID, amount, startTime, endTime)
}

// RequestStake is a paid mutator transaction binding the contract method 0x2dbf0344.
//
// Solidity: function requestStake(bytes publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime) returns()
func (_MpcCoordinator *MpcCoordinatorTransactorSession) RequestStake(publicKey []byte, nodeID string, amount *big.Int, startTime *big.Int, endTime *big.Int) (*types.Transaction, error) {
	return _MpcCoordinator.Contract.RequestStake(&_MpcCoordinator.TransactOpts, publicKey, nodeID, amount, startTime, endTime)
}

// MpcCoordinatorKeyGeneratedIterator is returned from FilterKeyGenerated and is used to iterate over the raw logs and unpacked data for KeyGenerated events raised by the MpcCoordinator contract.
type MpcCoordinatorKeyGeneratedIterator struct {
	Event *MpcCoordinatorKeyGenerated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorKeyGeneratedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorKeyGenerated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorKeyGenerated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorKeyGeneratedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorKeyGeneratedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorKeyGenerated represents a KeyGenerated event raised by the MpcCoordinator contract.
type MpcCoordinatorKeyGenerated struct {
	GroupId   [32]byte
	PublicKey []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterKeyGenerated is a free log retrieval operation binding the contract event 0x767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d.
//
// Solidity: event KeyGenerated(bytes32 indexed groupId, bytes publicKey)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterKeyGenerated(opts *bind.FilterOpts, groupId [][32]byte) (*MpcCoordinatorKeyGeneratedIterator, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "KeyGenerated", groupIdRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorKeyGeneratedIterator{contract: _MpcCoordinator.contract, event: "KeyGenerated", logs: logs, sub: sub}, nil
}

// WatchKeyGenerated is a free log subscription operation binding the contract event 0x767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d.
//
// Solidity: event KeyGenerated(bytes32 indexed groupId, bytes publicKey)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchKeyGenerated(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorKeyGenerated, groupId [][32]byte) (event.Subscription, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "KeyGenerated", groupIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorKeyGenerated)
				if err := _MpcCoordinator.contract.UnpackLog(event, "KeyGenerated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseKeyGenerated is a log parse operation binding the contract event 0x767b7aa89023ecd2db985822c15a32856d9106f50b5b2d5a65aa0f30d3cf457d.
//
// Solidity: event KeyGenerated(bytes32 indexed groupId, bytes publicKey)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseKeyGenerated(log types.Log) (*MpcCoordinatorKeyGenerated, error) {
	event := new(MpcCoordinatorKeyGenerated)
	if err := _MpcCoordinator.contract.UnpackLog(event, "KeyGenerated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorKeygenRequestAddedIterator is returned from FilterKeygenRequestAdded and is used to iterate over the raw logs and unpacked data for KeygenRequestAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorKeygenRequestAddedIterator struct {
	Event *MpcCoordinatorKeygenRequestAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorKeygenRequestAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorKeygenRequestAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorKeygenRequestAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorKeygenRequestAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorKeygenRequestAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorKeygenRequestAdded represents a KeygenRequestAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorKeygenRequestAdded struct {
	GroupId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterKeygenRequestAdded is a free log retrieval operation binding the contract event 0x5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa7327.
//
// Solidity: event KeygenRequestAdded(bytes32 indexed groupId)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterKeygenRequestAdded(opts *bind.FilterOpts, groupId [][32]byte) (*MpcCoordinatorKeygenRequestAddedIterator, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "KeygenRequestAdded", groupIdRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorKeygenRequestAddedIterator{contract: _MpcCoordinator.contract, event: "KeygenRequestAdded", logs: logs, sub: sub}, nil
}

// WatchKeygenRequestAdded is a free log subscription operation binding the contract event 0x5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa7327.
//
// Solidity: event KeygenRequestAdded(bytes32 indexed groupId)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchKeygenRequestAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorKeygenRequestAdded, groupId [][32]byte) (event.Subscription, error) {

	var groupIdRule []interface{}
	for _, groupIdItem := range groupId {
		groupIdRule = append(groupIdRule, groupIdItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "KeygenRequestAdded", groupIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorKeygenRequestAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "KeygenRequestAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseKeygenRequestAdded is a log parse operation binding the contract event 0x5e169d3e7bcbd6275f0072b5b8ebc2971595796ad9715cabd718a8237baa7327.
//
// Solidity: event KeygenRequestAdded(bytes32 indexed groupId)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseKeygenRequestAdded(log types.Log) (*MpcCoordinatorKeygenRequestAdded, error) {
	event := new(MpcCoordinatorKeygenRequestAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "KeygenRequestAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorParticipantAddedIterator is returned from FilterParticipantAdded and is used to iterate over the raw logs and unpacked data for ParticipantAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorParticipantAddedIterator struct {
	Event *MpcCoordinatorParticipantAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorParticipantAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorParticipantAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorParticipantAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorParticipantAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorParticipantAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorParticipantAdded represents a ParticipantAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorParticipantAdded struct {
	PublicKey common.Hash
	GroupId   [32]byte
	Index     *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterParticipantAdded is a free log retrieval operation binding the contract event 0x39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a.
//
// Solidity: event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterParticipantAdded(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorParticipantAddedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "ParticipantAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorParticipantAddedIterator{contract: _MpcCoordinator.contract, event: "ParticipantAdded", logs: logs, sub: sub}, nil
}

// WatchParticipantAdded is a free log subscription operation binding the contract event 0x39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a.
//
// Solidity: event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchParticipantAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorParticipantAdded, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "ParticipantAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorParticipantAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "ParticipantAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseParticipantAdded is a log parse operation binding the contract event 0x39f1368dd39c286ea788ed1ca8b79dddbdad29f340f0100a5f2a60bd4d2f269a.
//
// Solidity: event ParticipantAdded(bytes indexed publicKey, bytes32 groupId, uint256 index)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseParticipantAdded(log types.Log) (*MpcCoordinatorParticipantAdded, error) {
	event := new(MpcCoordinatorParticipantAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "ParticipantAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorSignRequestAddedIterator is returned from FilterSignRequestAdded and is used to iterate over the raw logs and unpacked data for SignRequestAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestAddedIterator struct {
	Event *MpcCoordinatorSignRequestAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorSignRequestAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorSignRequestAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorSignRequestAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorSignRequestAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorSignRequestAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorSignRequestAdded represents a SignRequestAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestAdded struct {
	RequestId *big.Int
	PublicKey common.Hash
	Message   []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSignRequestAdded is a free log retrieval operation binding the contract event 0xfd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b22628.
//
// Solidity: event SignRequestAdded(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterSignRequestAdded(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorSignRequestAddedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "SignRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorSignRequestAddedIterator{contract: _MpcCoordinator.contract, event: "SignRequestAdded", logs: logs, sub: sub}, nil
}

// WatchSignRequestAdded is a free log subscription operation binding the contract event 0xfd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b22628.
//
// Solidity: event SignRequestAdded(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchSignRequestAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorSignRequestAdded, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "SignRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorSignRequestAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignRequestAdded is a log parse operation binding the contract event 0xfd47ace1305a71239c6719afa87da2a0b202b0d7d727aad7f69ad1a934b22628.
//
// Solidity: event SignRequestAdded(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseSignRequestAdded(log types.Log) (*MpcCoordinatorSignRequestAdded, error) {
	event := new(MpcCoordinatorSignRequestAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorSignRequestStartedIterator is returned from FilterSignRequestStarted and is used to iterate over the raw logs and unpacked data for SignRequestStarted events raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestStartedIterator struct {
	Event *MpcCoordinatorSignRequestStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorSignRequestStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorSignRequestStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorSignRequestStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorSignRequestStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorSignRequestStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorSignRequestStarted represents a SignRequestStarted event raised by the MpcCoordinator contract.
type MpcCoordinatorSignRequestStarted struct {
	RequestId *big.Int
	PublicKey common.Hash
	Message   []byte
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSignRequestStarted is a free log retrieval operation binding the contract event 0x279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad303.
//
// Solidity: event SignRequestStarted(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterSignRequestStarted(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorSignRequestStartedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "SignRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorSignRequestStartedIterator{contract: _MpcCoordinator.contract, event: "SignRequestStarted", logs: logs, sub: sub}, nil
}

// WatchSignRequestStarted is a free log subscription operation binding the contract event 0x279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad303.
//
// Solidity: event SignRequestStarted(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchSignRequestStarted(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorSignRequestStarted, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "SignRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorSignRequestStarted)
				if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignRequestStarted is a log parse operation binding the contract event 0x279ae2c17b7204cd61039a5a8ea3db27acc71416ea84fb62e95335c8b24ad303.
//
// Solidity: event SignRequestStarted(uint256 requestId, bytes indexed publicKey, bytes message)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseSignRequestStarted(log types.Log) (*MpcCoordinatorSignRequestStarted, error) {
	event := new(MpcCoordinatorSignRequestStarted)
	if err := _MpcCoordinator.contract.UnpackLog(event, "SignRequestStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorStakeRequestAddedIterator is returned from FilterStakeRequestAdded and is used to iterate over the raw logs and unpacked data for StakeRequestAdded events raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestAddedIterator struct {
	Event *MpcCoordinatorStakeRequestAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorStakeRequestAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorStakeRequestAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorStakeRequestAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorStakeRequestAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorStakeRequestAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorStakeRequestAdded represents a StakeRequestAdded event raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestAdded struct {
	RequestId *big.Int
	PublicKey common.Hash
	NodeID    string
	Amount    *big.Int
	StartTime *big.Int
	EndTime   *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStakeRequestAdded is a free log retrieval operation binding the contract event 0x18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594.
//
// Solidity: event StakeRequestAdded(uint256 requestId, bytes indexed publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterStakeRequestAdded(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorStakeRequestAddedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "StakeRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorStakeRequestAddedIterator{contract: _MpcCoordinator.contract, event: "StakeRequestAdded", logs: logs, sub: sub}, nil
}

// WatchStakeRequestAdded is a free log subscription operation binding the contract event 0x18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594.
//
// Solidity: event StakeRequestAdded(uint256 requestId, bytes indexed publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchStakeRequestAdded(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorStakeRequestAdded, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "StakeRequestAdded", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorStakeRequestAdded)
				if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeRequestAdded is a log parse operation binding the contract event 0x18d59ead2751a952ffa140860eedfe61eefb762649f64d9a222b9c8e2b7bf594.
//
// Solidity: event StakeRequestAdded(uint256 requestId, bytes indexed publicKey, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseStakeRequestAdded(log types.Log) (*MpcCoordinatorStakeRequestAdded, error) {
	event := new(MpcCoordinatorStakeRequestAdded)
	if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MpcCoordinatorStakeRequestStartedIterator is returned from FilterStakeRequestStarted and is used to iterate over the raw logs and unpacked data for StakeRequestStarted events raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestStartedIterator struct {
	Event *MpcCoordinatorStakeRequestStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MpcCoordinatorStakeRequestStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MpcCoordinatorStakeRequestStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MpcCoordinatorStakeRequestStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MpcCoordinatorStakeRequestStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MpcCoordinatorStakeRequestStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MpcCoordinatorStakeRequestStarted represents a StakeRequestStarted event raised by the MpcCoordinator contract.
type MpcCoordinatorStakeRequestStarted struct {
	RequestId          *big.Int
	PublicKey          common.Hash
	ParticipantIndices []*big.Int
	NodeID             string
	Amount             *big.Int
	StartTime          *big.Int
	EndTime            *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterStakeRequestStarted is a free log retrieval operation binding the contract event 0x288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f30486.
//
// Solidity: event StakeRequestStarted(uint256 requestId, bytes indexed publicKey, uint256[] participantIndices, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) FilterStakeRequestStarted(opts *bind.FilterOpts, publicKey [][]byte) (*MpcCoordinatorStakeRequestStartedIterator, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.FilterLogs(opts, "StakeRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return &MpcCoordinatorStakeRequestStartedIterator{contract: _MpcCoordinator.contract, event: "StakeRequestStarted", logs: logs, sub: sub}, nil
}

// WatchStakeRequestStarted is a free log subscription operation binding the contract event 0x288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f30486.
//
// Solidity: event StakeRequestStarted(uint256 requestId, bytes indexed publicKey, uint256[] participantIndices, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) WatchStakeRequestStarted(opts *bind.WatchOpts, sink chan<- *MpcCoordinatorStakeRequestStarted, publicKey [][]byte) (event.Subscription, error) {

	var publicKeyRule []interface{}
	for _, publicKeyItem := range publicKey {
		publicKeyRule = append(publicKeyRule, publicKeyItem)
	}

	logs, sub, err := _MpcCoordinator.contract.WatchLogs(opts, "StakeRequestStarted", publicKeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MpcCoordinatorStakeRequestStarted)
				if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeRequestStarted is a log parse operation binding the contract event 0x288b3cb79b7b3694315e9132713d254471d922b469ac4c7f26fee7fe49f30486.
//
// Solidity: event StakeRequestStarted(uint256 requestId, bytes indexed publicKey, uint256[] participantIndices, string nodeID, uint256 amount, uint256 startTime, uint256 endTime)
func (_MpcCoordinator *MpcCoordinatorFilterer) ParseStakeRequestStarted(log types.Log) (*MpcCoordinatorStakeRequestStarted, error) {
	event := new(MpcCoordinatorStakeRequestStarted)
	if err := _MpcCoordinator.contract.UnpackLog(event, "StakeRequestStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
