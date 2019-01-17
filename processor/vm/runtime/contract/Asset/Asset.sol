pragma solidity ^0.4.24;

contract Asset { 
    uint256 public totalSupply;

    function Asset() {
        totalSupply = 10;
    }

    function reg(string desc) public {
        issueasset(desc);
    }
    function add(address assetId, uint256 value) public {
        addasset(assetId,value);
    }
    function transAsset(address to, address assetId, uint256 value) public payable {
        to.transferex(assetId, value);
    }
    function setname(address newOwner, address assetId) public {
        setassetowner(assetId, newOwner);
    }
    function getbalance(address to, address assetId) public returns(uint) {
        return to.balanceex(assetId);
    }
}
