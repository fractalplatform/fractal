pragma solidity ^0.4.24;

contract WithdrawFee {
    constructor() public payable {
    }

    function withdrawAssetFee(uint256 assetId) public payable {
        withdrawfee(assetId, 2);
    }

    function withdrawAccountFee(uint256 userId) public payable {
        withdrawfee(userId, 1);
    }
}