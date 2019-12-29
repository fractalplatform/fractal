pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

contract AssetAPI {
    function IssueAsset(string name ,string symbol ,uint256 amount, uint64 decimals, string founder, string owner, uint256 limit, string desc) external;
    function IncreaseAsset(string to, uint64 assetID, uint256 amount) external;
}

contract TestAsset {
    AssetAPI asset = AssetAPI(address(bytes20("fractalasset")));

    function testIssueAsset(string name ,string symbol ,uint256 amount, uint64 decimals, string founder, string owner, uint256 limit, string desc) public {
        asset.IssueAsset(name, symbol, amount, decimals, founder, owner, limit, desc);
    }

    function testIncreaseAsset(string to, uint64 assetID, uint256 amount) public {
        asset.IncreaseAsset(to, assetID, amount);
    }
}