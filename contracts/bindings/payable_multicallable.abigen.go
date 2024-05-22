// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

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
	_ = abi.ConvertType
)

// PayableMulticallableMetaData contains all meta data concerning the PayableMulticallable contract.
var PayableMulticallableMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"multicall\",\"inputs\":[{\"name\":\"requireSuccess\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"data\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"stateMutability\":\"payable\"},{\"type\":\"error\",\"name\":\"AmountOverflow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EthTransferFailed\",\"inputs\":[]}]",
}

// PayableMulticallableABI is the input ABI used to generate the binding from.
// Deprecated: Use PayableMulticallableMetaData.ABI instead.
var PayableMulticallableABI = PayableMulticallableMetaData.ABI

// PayableMulticallable is an auto generated Go binding around an Ethereum contract.
type PayableMulticallable struct {
	PayableMulticallableCaller     // Read-only binding to the contract
	PayableMulticallableTransactor // Write-only binding to the contract
	PayableMulticallableFilterer   // Log filterer for contract events
}

// PayableMulticallableCaller is an auto generated read-only Go binding around an Ethereum contract.
type PayableMulticallableCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayableMulticallableTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PayableMulticallableTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayableMulticallableFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PayableMulticallableFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayableMulticallableSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PayableMulticallableSession struct {
	Contract     *PayableMulticallable // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// PayableMulticallableCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PayableMulticallableCallerSession struct {
	Contract *PayableMulticallableCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// PayableMulticallableTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PayableMulticallableTransactorSession struct {
	Contract     *PayableMulticallableTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// PayableMulticallableRaw is an auto generated low-level Go binding around an Ethereum contract.
type PayableMulticallableRaw struct {
	Contract *PayableMulticallable // Generic contract binding to access the raw methods on
}

// PayableMulticallableCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PayableMulticallableCallerRaw struct {
	Contract *PayableMulticallableCaller // Generic read-only contract binding to access the raw methods on
}

// PayableMulticallableTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PayableMulticallableTransactorRaw struct {
	Contract *PayableMulticallableTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPayableMulticallable creates a new instance of PayableMulticallable, bound to a specific deployed contract.
func NewPayableMulticallable(address common.Address, backend bind.ContractBackend) (*PayableMulticallable, error) {
	contract, err := bindPayableMulticallable(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PayableMulticallable{PayableMulticallableCaller: PayableMulticallableCaller{contract: contract}, PayableMulticallableTransactor: PayableMulticallableTransactor{contract: contract}, PayableMulticallableFilterer: PayableMulticallableFilterer{contract: contract}}, nil
}

// NewPayableMulticallableCaller creates a new read-only instance of PayableMulticallable, bound to a specific deployed contract.
func NewPayableMulticallableCaller(address common.Address, caller bind.ContractCaller) (*PayableMulticallableCaller, error) {
	contract, err := bindPayableMulticallable(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PayableMulticallableCaller{contract: contract}, nil
}

// NewPayableMulticallableTransactor creates a new write-only instance of PayableMulticallable, bound to a specific deployed contract.
func NewPayableMulticallableTransactor(address common.Address, transactor bind.ContractTransactor) (*PayableMulticallableTransactor, error) {
	contract, err := bindPayableMulticallable(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PayableMulticallableTransactor{contract: contract}, nil
}

// NewPayableMulticallableFilterer creates a new log filterer instance of PayableMulticallable, bound to a specific deployed contract.
func NewPayableMulticallableFilterer(address common.Address, filterer bind.ContractFilterer) (*PayableMulticallableFilterer, error) {
	contract, err := bindPayableMulticallable(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PayableMulticallableFilterer{contract: contract}, nil
}

// bindPayableMulticallable binds a generic wrapper to an already deployed contract.
func bindPayableMulticallable(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PayableMulticallableMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayableMulticallable *PayableMulticallableRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayableMulticallable.Contract.PayableMulticallableCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayableMulticallable *PayableMulticallableRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayableMulticallable.Contract.PayableMulticallableTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayableMulticallable *PayableMulticallableRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayableMulticallable.Contract.PayableMulticallableTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayableMulticallable *PayableMulticallableCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayableMulticallable.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayableMulticallable *PayableMulticallableTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayableMulticallable.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayableMulticallable *PayableMulticallableTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayableMulticallable.Contract.contract.Transact(opts, method, params...)
}

// Multicall is a paid mutator transaction binding the contract method 0xafe7260f.
//
// Solidity: function multicall(bool requireSuccess, bytes[] data) payable returns(bytes[])
func (_PayableMulticallable *PayableMulticallableTransactor) Multicall(opts *bind.TransactOpts, requireSuccess bool, data [][]byte) (*types.Transaction, error) {
	return _PayableMulticallable.contract.Transact(opts, "multicall", requireSuccess, data)
}

// Multicall is a paid mutator transaction binding the contract method 0xafe7260f.
//
// Solidity: function multicall(bool requireSuccess, bytes[] data) payable returns(bytes[])
func (_PayableMulticallable *PayableMulticallableSession) Multicall(requireSuccess bool, data [][]byte) (*types.Transaction, error) {
	return _PayableMulticallable.Contract.Multicall(&_PayableMulticallable.TransactOpts, requireSuccess, data)
}

// Multicall is a paid mutator transaction binding the contract method 0xafe7260f.
//
// Solidity: function multicall(bool requireSuccess, bytes[] data) payable returns(bytes[])
func (_PayableMulticallable *PayableMulticallableTransactorSession) Multicall(requireSuccess bool, data [][]byte) (*types.Transaction, error) {
	return _PayableMulticallable.Contract.Multicall(&_PayableMulticallable.TransactOpts, requireSuccess, data)
}
