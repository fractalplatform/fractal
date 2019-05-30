pragma solidity ^0.4.24;

contract Asset {
    constructor() public payable {
    }
    function issue(string desc) public{
        uint256 id;
        id = issueasset(desc);
        log1(bytes32(id), "id");
    } 
    function increase(uint256 assetId, address to, uint256 value ) public {
        addasset(assetId,to,value);
    }
    function transfer(uint256 assetId, address to, uint256 value) public  {
        to.transfer(assetId, value);
    }
    function changeowner(address newOwner, uint256 assetId) public {
        setassetowner(assetId, newOwner);
    }
    function destroy(uint256 assetId, uint256 value) public {
        uint256 d;
        d = destroyasset(assetId, value);
        log1(bytes32(d), "value");
    }
}