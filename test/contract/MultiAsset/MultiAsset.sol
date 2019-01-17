pragma solidity ^0.4.24;

contract MultiAsset { 
    function() payable {

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
    function changeOwner(address newOwner, address assetId) public {
        setassetowner(assetId, newOwner);
    }
}
