pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;
import "plugin.sol";

contract ConsensusAPI is Plugin {
    address constant consensus = address(bytes20("fractaldpos"));
    struct MinerInfo {
        address OwnerAccount;
        address SignAccount;
        uint256 RegisterNumber;
        uint256 Weight;
        uint256 Balance;
        uint256 Epoch;
    }
    function GetMinerInfo(address miner) external returns(MinerInfo memory){
        Plugin.Call(consensus);
    }

    function UnregisterMiner() public {
        Plugin.Call(consensus);
    }
    function RegisterMiner(address miner) external payable {
        Plugin.Call(consensus);
    }

    event InfoLog(address);
    function testRead(address miner) public {
        MinerInfo memory info = this.GetMinerInfo(miner);
        emit InfoLog(info.OwnerAccount);
    }
}