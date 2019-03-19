pragma solidity ^0.4.24;

contract MultiAsset {
    constructor() public payable {

    }

    function reg(string desc) public returns(uint256){
        return issueasset(desc);
    }

    function destroyasset(uint256 assetId, uint256 value) public returns(uint256)  {
        return destroyasset(assetId, value);
    }

    function getaccountid(address desc) public returns(uint256) {
        return getaccountid(desc);
    }

    function add(uint256 assetId, address toname,uint256 value) public returns(uint256)  {
        return addasset(assetId,toname,value);
    }

    function changeOwner(address newOwner, uint256 assetId) public returns(uint256)  {
       return setassetowner(assetId, newOwner);
    }
    
    function transAsset(address to, uint256 assetId, uint256 value) public payable {
        to.transferex(assetId, value);
    }
    
    function getBalanceEx(address to,uint256 assetId) public {
        log1(bytes32(to.balanceex(assetId)),"getbalanceex");
    }
    
    function getAssetAmount(uint256 assetId, uint256 time) public{
        uint256 x;
        x = assetamount(assetId,time);
        log1(bytes32(x),"getassetamount");
    }
    
    function getSnapshotTime(uint256 t,uint256 time) public{
        uint256 x;
        x = snapshottime(t,time);
        log1(bytes32(x),"getSnapshotTime" ); 
    }
    function getSnapBalance(address to,uint256 assetId,uint256 time,uint256 t) public {
        uint256 x ;
        x = to.snapbalance(assetId,time,t);
        log1(bytes32(x),"getSnapBalance");
    }
}
