// SPDX-License-Identifier: MIT
pragma solidity >=0.8.19 <=0.9.0;

import {Script} from "forge-std/Script.sol";

abstract contract BaseScript is Script {
    /// @dev Included to enable compilation of the script without a $MNEMONIC
    /// environment variable.
    uint256 internal constant TEST_PRIVKEY =
        0xfffdbb37105441e14b0ee6330d855d8504ff39e705c3afa8f859ac9865f99306;

    /// @dev Needed for the deterministic deployments.
    bytes32 internal constant ZERO_SALT = bytes32(0);

    /// @dev The private key of the transaction broadcaster.
    uint256 internal broadcaster;

    /// @dev Initializes the transaction broadcaster like this:
    ///
    /// - If $ETH_FROM is defined, use it.
    /// - Otherwise, derive the broadcaster address from $MNEMONIC.
    /// - If $MNEMONIC is not defined, default to a test mnemonic.
    ///
    /// The use case for $ETH_FROM is to specify the broadcaster key and its
    /// address via the command line.
    constructor() {
        broadcaster =
            vm.envOr({name: "DEPLOYER_PRIVKEY", defaultValue: TEST_PRIVKEY});
    }

    modifier broadcast() {
        vm.startBroadcast(broadcaster);
        _;
        vm.stopBroadcast();
    }
}
