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

// PayableMulticallMetaData contains all meta data concerning the PayableMulticall contract.
var PayableMulticallMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"receive\",\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"incNumber\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"multicall\",\"inputs\":[{\"name\":\"requireSuccess\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"data\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes[]\",\"internalType\":\"bytes[]\"}],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"multicallBalance\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"error\",\"name\":\"AmountOverflow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"EthTransferFailed\",\"inputs\":[]}]",
}

// PayableMulticallABI is the input ABI used to generate the binding from.
// Deprecated: Use PayableMulticallMetaData.ABI instead.
var PayableMulticallABI = PayableMulticallMetaData.ABI

// PayableMulticall is an auto generated Go binding around an Ethereum contract.
type PayableMulticall struct {
	PayableMulticallCaller     // Read-only binding to the contract
	PayableMulticallTransactor // Write-only binding to the contract
	PayableMulticallFilterer   // Log filterer for contract events
}

// PayableMulticallCaller is an auto generated read-only Go binding around an Ethereum contract.
type PayableMulticallCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayableMulticallTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PayableMulticallTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayableMulticallFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PayableMulticallFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PayableMulticallSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PayableMulticallSession struct {
	Contract     *PayableMulticall // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PayableMulticallCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PayableMulticallCallerSession struct {
	Contract *PayableMulticallCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// PayableMulticallTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PayableMulticallTransactorSession struct {
	Contract     *PayableMulticallTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// PayableMulticallRaw is an auto generated low-level Go binding around an Ethereum contract.
type PayableMulticallRaw struct {
	Contract *PayableMulticall // Generic contract binding to access the raw methods on
}

// PayableMulticallCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PayableMulticallCallerRaw struct {
	Contract *PayableMulticallCaller // Generic read-only contract binding to access the raw methods on
}

// PayableMulticallTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PayableMulticallTransactorRaw struct {
	Contract *PayableMulticallTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPayableMulticall creates a new instance of PayableMulticall, bound to a specific deployed contract.
func NewPayableMulticall(address common.Address, backend bind.ContractBackend) (*PayableMulticall, error) {
	contract, err := bindPayableMulticall(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PayableMulticall{PayableMulticallCaller: PayableMulticallCaller{contract: contract}, PayableMulticallTransactor: PayableMulticallTransactor{contract: contract}, PayableMulticallFilterer: PayableMulticallFilterer{contract: contract}}, nil
}

// NewPayableMulticallCaller creates a new read-only instance of PayableMulticall, bound to a specific deployed contract.
func NewPayableMulticallCaller(address common.Address, caller bind.ContractCaller) (*PayableMulticallCaller, error) {
	contract, err := bindPayableMulticall(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PayableMulticallCaller{contract: contract}, nil
}

// NewPayableMulticallTransactor creates a new write-only instance of PayableMulticall, bound to a specific deployed contract.
func NewPayableMulticallTransactor(address common.Address, transactor bind.ContractTransactor) (*PayableMulticallTransactor, error) {
	contract, err := bindPayableMulticall(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PayableMulticallTransactor{contract: contract}, nil
}

// NewPayableMulticallFilterer creates a new log filterer instance of PayableMulticall, bound to a specific deployed contract.
func NewPayableMulticallFilterer(address common.Address, filterer bind.ContractFilterer) (*PayableMulticallFilterer, error) {
	contract, err := bindPayableMulticall(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PayableMulticallFilterer{contract: contract}, nil
}

// bindPayableMulticall binds a generic wrapper to an already deployed contract.
func bindPayableMulticall(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PayableMulticallMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayableMulticall *PayableMulticallRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayableMulticall.Contract.PayableMulticallCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayableMulticall *PayableMulticallRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayableMulticall.Contract.PayableMulticallTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayableMulticall *PayableMulticallRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayableMulticall.Contract.PayableMulticallTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PayableMulticall *PayableMulticallCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PayableMulticall.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PayableMulticall *PayableMulticallTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayableMulticall.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PayableMulticall *PayableMulticallTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PayableMulticall.Contract.contract.Transact(opts, method, params...)
}

// MulticallBalance is a free data retrieval call binding the contract method 0x2dd38eaa.
//
// Solidity: function multicallBalance() view returns(uint256)
func (_PayableMulticall *PayableMulticallCaller) MulticallBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PayableMulticall.contract.Call(opts, &out, "multicallBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MulticallBalance is a free data retrieval call binding the contract method 0x2dd38eaa.
//
// Solidity: function multicallBalance() view returns(uint256)
func (_PayableMulticall *PayableMulticallSession) MulticallBalance() (*big.Int, error) {
	return _PayableMulticall.Contract.MulticallBalance(&_PayableMulticall.CallOpts)
}

// MulticallBalance is a free data retrieval call binding the contract method 0x2dd38eaa.
//
// Solidity: function multicallBalance() view returns(uint256)
func (_PayableMulticall *PayableMulticallCallerSession) MulticallBalance() (*big.Int, error) {
	return _PayableMulticall.Contract.MulticallBalance(&_PayableMulticall.CallOpts)
}

// IncNumber is a paid mutator transaction binding the contract method 0xafd97196.
//
// Solidity: function incNumber(uint256 amount) payable returns(uint256)
func (_PayableMulticall *PayableMulticallTransactor) IncNumber(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _PayableMulticall.contract.Transact(opts, "incNumber", amount)
}

// IncNumber is a paid mutator transaction binding the contract method 0xafd97196.
//
// Solidity: function incNumber(uint256 amount) payable returns(uint256)
func (_PayableMulticall *PayableMulticallSession) IncNumber(amount *big.Int) (*types.Transaction, error) {
	return _PayableMulticall.Contract.IncNumber(&_PayableMulticall.TransactOpts, amount)
}

// IncNumber is a paid mutator transaction binding the contract method 0xafd97196.
//
// Solidity: function incNumber(uint256 amount) payable returns(uint256)
func (_PayableMulticall *PayableMulticallTransactorSession) IncNumber(amount *big.Int) (*types.Transaction, error) {
	return _PayableMulticall.Contract.IncNumber(&_PayableMulticall.TransactOpts, amount)
}

// Multicall is a paid mutator transaction binding the contract method 0xafe7260f.
//
// Solidity: function multicall(bool requireSuccess, bytes[] data) payable returns(bytes[])
func (_PayableMulticall *PayableMulticallTransactor) Multicall(opts *bind.TransactOpts, requireSuccess bool, data [][]byte) (*types.Transaction, error) {
	return _PayableMulticall.contract.Transact(opts, "multicall", requireSuccess, data)
}

// Multicall is a paid mutator transaction binding the contract method 0xafe7260f.
//
// Solidity: function multicall(bool requireSuccess, bytes[] data) payable returns(bytes[])
func (_PayableMulticall *PayableMulticallSession) Multicall(requireSuccess bool, data [][]byte) (*types.Transaction, error) {
	return _PayableMulticall.Contract.Multicall(&_PayableMulticall.TransactOpts, requireSuccess, data)
}

// Multicall is a paid mutator transaction binding the contract method 0xafe7260f.
//
// Solidity: function multicall(bool requireSuccess, bytes[] data) payable returns(bytes[])
func (_PayableMulticall *PayableMulticallTransactorSession) Multicall(requireSuccess bool, data [][]byte) (*types.Transaction, error) {
	return _PayableMulticall.Contract.Multicall(&_PayableMulticall.TransactOpts, requireSuccess, data)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_PayableMulticall *PayableMulticallTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PayableMulticall.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_PayableMulticall *PayableMulticallSession) Receive() (*types.Transaction, error) {
	return _PayableMulticall.Contract.Receive(&_PayableMulticall.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_PayableMulticall *PayableMulticallTransactorSession) Receive() (*types.Transaction, error) {
	return _PayableMulticall.Contract.Receive(&_PayableMulticall.TransactOpts)
}
