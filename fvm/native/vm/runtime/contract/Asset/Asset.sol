pragma solidity ^0.4.24;

contract Asset {
    uint256 public totalSupply;

    constructor() public {
        totalSupply = 10;
    }

    function reg(string desc) public returns(uint256){
        return issueasset(desc);
    }
    function add(uint assetId, address to, uint256 value) public {
        addasset(assetId,to,value);
    }
    function getAssetId() public returns (uint256) {
        return msg.assetid;
    }
    function transAsset(address to, uint assetId, uint256 value) public payable {
        to.transfer(assetId, value);
    }
    function setname(address newOwner, uint assetId) public {
        setassetowner(assetId, newOwner);
    }
    function getbalance(address to, uint assetId) public returns(uint256) {
        return to.balanceex(assetId);
    }
    // function getAssetAmount(uint256 assetId, uint256 t) public returns (uint256){
    //     return assetamount(assetId,t);
    // }
    function getSnapshotTime(uint256 i,uint256 t) public returns (uint256){
        return snapshottime(i,t);
    }
    function getSnapBalance(address to,uint256 assetId,uint256 t) public returns (uint256){
        return to.snapbalance(assetId,t,0);
    }
}
