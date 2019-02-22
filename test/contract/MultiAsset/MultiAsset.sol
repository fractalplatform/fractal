pragma solidity ^0.4.24;

contract MultiAsset {
    function() public payable {
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
   function getBalanceEx(address to,address assetId) public {
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
    function getSnapBalance(address to,uint256 assetId,uint256 time) public {
        uint256 x ;
        x = to.snapbalance(assetId,time);
        log1(bytes32(x),"getSnapBalance");
    }
}
