// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "../lib/forge-std/src/Script.sol";
import "../src/ToWatch.sol";

contract DeployScript is Script {
    function run() external returns (ToWatch) {
        vm.startBroadcast(0xfffdbb37105441e14b0ee6330d855d8504ff39e705c3afa8f859ac9865f99306);
        ToWatch x = new ToWatch();
        vm.stopBroadcast();
        return x;
    }
}