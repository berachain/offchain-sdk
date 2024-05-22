package bindings

// Utils
//
//go:generate abigen --pkg bindings --abi ./out/Multicall3.sol/Multicall3.abi.json --out bindings/multicall_3.abigen.go --type Multicall3
//go:generate abigen --pkg bindings --abi ./out/IERC20.sol/IERC20.abi.json --out bindings/erc20.abigen.go --type IERC20

//go:generate abigen --pkg bindings --abi ./out/PayableMulticall.sol/PayableMulticall.abi.json --out bindings/payable_multicall.abigen.go --type PayableMulticall
//go:generate abigen --pkg bindings --abi ./out/PayableMulticallable.sol/PayableMulticallable.abi.json --out bindings/payable_multicallable.abigen.go --type PayableMulticallable
