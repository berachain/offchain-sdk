// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "../lib/forge-std/src/Script.sol";
import "../src/ToListen.sol";

/// @notice For the examples/listener project.
contract DeployScript is Script {
    function run() external returns (ToListen) {
        vm.startBroadcast(
            0xfffdbb37105441e14b0ee6330d855d8504ff39e705c3afa8f859ac9865f99306
        );
        ToListen x = new ToListen();
        vm.stopBroadcast();
        return x;
    }
}
