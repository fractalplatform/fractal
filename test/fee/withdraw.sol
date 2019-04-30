pragma solidity ^0.4.24;

contract WithdrawFee {
    constructor() public payable {
    }

    function withdrawAssetFee(uint256 assetId) public payable {
        withdrawfee(assetId, 0);
    }

    function withdrawContractFee(uint256 userId) public payable {
        withdrawfee(userId, 1);
    }

    function withdrawCoinbaseFee(uint256 userId) public payable {
        withdrawfee(userId, 2);
    }
}