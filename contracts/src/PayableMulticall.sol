// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { PayableMulticallable } from  "../lib/transient-goodies/src/PayableMulticallable.sol";

contract PayableMulticall is PayableMulticallable {
    uint256 public multicallBalance;

    constructor() PayableMulticallable() {}

    /// @notice is multicallable! increments number by how much is payed to the contract.
    /// @return multicallBalance after increment
    function incNumber() external payable standalonePayable returns (uint256) {
        if (msg.value == 0) revert();
        multicallBalance += useValue(msg.value);
        return multicallBalance;
    }

    receive() external payable {}
}