// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.19;

import "../src/TeamspeakAccountRegistry.sol";
import "../src/TeamspeakAccountProxy.sol";
import "../src/TeamspeakAccount.sol";
import "../src/TeamspeakAccountToken.sol";
import "forge-std/console2.sol";

import {BaseScript} from "./Base.s.sol";

/// @title Deployment and Registration Contract
/// @notice This contract deploys and registers various contracts for Teamspeak
/// accounts and NFTs.
/// @dev This contract serves as a deployment script to create and configure
/// necessary contracts.
contract Deploy is BaseScript {
    /// @notice The TeamspeakAccountRegistry contract that stores account
    /// information.
    TeamspeakAccountRegistry public registry;

    /// @notice The TeamspeakAccount contract that defines the implementation of
    /// accounts.
    TeamspeakAccount public implementation;

    /// @notice The TeamspeakAccountProxy contract that acts as a proxy for
    /// account functionality.
    TeamspeakAccountProxy public proxy;

    /// @notice The TeamspeakAccountToken contract that represents NFTs
    /// associated with accounts.
    TeamspeakAccountToken public nft;

    /// @notice Executes the deployment process.
    function run() public {
        // Deploy all the contracts.
        deployContracts();

        // Register an account for the deployer.
        registerAccount(address(0x20f33CE90A13a4b5E7697E3544c3083B8F8A51D4));

        // Display all user accounts.
        address[] memory userAccounts =
            getUserAccounts(0x20f33CE90A13a4b5E7697E3544c3083B8F8A51D4);

        for (uint256 i = 0; i < userAccounts.length; i++) {
            console2.log("User account #%d: %s", i, userAccounts[i]);
        }
    }

    /*//////////////////////////////////////////////////////////////
                                DEPLOYMENT
    //////////////////////////////////////////////////////////////*/

    /// @notice Deploys all necessary contracts for Teamspeak accounts and NFTs.
    function deployContracts() public broadcast {
        // Start broadcasting deployment steps.
        // vm.startBroadcast(broadcaster);

        // Deploy the TeamspeakAccountToken contract.
        nft = new TeamspeakAccountToken();

        // Deploy the TeamspeakAccountRegistry contract.
        registry = new TeamspeakAccountRegistry();

        // Deploy the TeamspeakAccount contract.
        implementation = new TeamspeakAccount();

        // Deploy the TeamspeakAccountProxy contract, connected to the
        // implementation.
        proxy = new TeamspeakAccountProxy(address(implementation));

        // Log the addresses of the deployed contracts.
        console2.log("Deployed TeamspeakAccountToken at %s", address(nft));
        console2.log(
            "Deployed TeamspeakAccountRegistry at %s", address(registry)
        );
        console2.log("Deployed TeamspeakAccount at %s", address(implementation));
        console2.log("Deployed TeamspeakAccountProxy at %s", address(proxy));
    }

    /*//////////////////////////////////////////////////////////////
                                  UTILS
    //////////////////////////////////////////////////////////////*/

    /// @notice Registers a Teamspeak account for a given user address.
    /// @param user The user's Ethereum address.
    /// @return tsAccount The created TeamspeakAccount instance.
    function registerAccount(address user)
        public
        broadcast
        returns (TeamspeakAccount tsAccount)
    {
        // Get the total supply of NFTs (tokenIDs).
        uint256 nextTokenID = nft.totalSupply();

        // Mint an NFT to the user.
        nft.mint(user);

        // Create a new Teamspeak account tied to the NFT.
        tsAccount = TeamspeakAccount(
            payable(
                registry.createAccount(
                    address(proxy),
                    block.chainid,
                    address(nft),
                    nextTokenID,
                    uint256(ZERO_SALT),
                    ""
                )
            )
        );

        console2.log(
            "Registered account %s for user %s, with tokenID %s",
            address(tsAccount),
            user,
            nextTokenID
        );
    }

    /*//////////////////////////////////////////////////////////////
                                  VIEWS
    //////////////////////////////////////////////////////////////*/

    /**
     * @notice Get the accounts associated with the NFTs owned by a user.
     * @param user The address of the user whose NFTs' accounts need to be
     * retrieved.
     * @return An array of addresses representing the accounts associated with
     * the user's NFTs.
     */
    function getUserAccounts(address user)
        public
        view
        returns (address[] memory)
    {
        // See how many NFTs the user has.
        uint256 ownedTokens = nft.balanceOf(user);

        // Iterate through and collect all the accounts associated with the NFTs
        // that the user has.
        address[] memory accounts = new address[](ownedTokens);
        unchecked {
            for (uint256 i = 0; i < ownedTokens; i++) {
                uint256 tokenID = nft.tokenOfOwnerByIndex(user, i);
                accounts[i] = registry.account(
                    address(proxy),
                    block.chainid,
                    address(nft),
                    tokenID,
                    uint256(ZERO_SALT)
                );
            }
        }

        // Return the accounts.
        return accounts;
    }
}
