pragma solidity ^0.4.24;

contract MultiAsset {
    uint256 assetID = 1;
    constructor() public payable {
    }
    function reg(string desc) public payable{
        assetID = issueasset(desc);
    } 
    function add(uint256 assetId, address to, uint256 value ) public {
        addasset(assetId,to,value);
    }
    function transAsset(address to, uint256 value) public payable {
        to.transfer(assetID, value);
    }
    function changeOwner(address newOwner, uint256 assetId) public {
        setassetowner(assetId, newOwner);
    }
}
