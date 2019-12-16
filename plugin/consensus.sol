pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

interface ConsensusAPI {
    struct MinerInfo {
        address OwnerAccount;
        address SignAccount;
        uint256 RegisterNumber;
        uint256 Weight;
        uint256 Balance;
        uint256 Epoch;
    }
    function GetMinerInfo(address miner) external returns(MinerInfo memory);
    function UnregisterMiner() external;
    function RegisterMiner(address miner) external payable;
}

contract TestRead {
    ConsensusAPI constant consensus = ConsensusAPI(address(bytes20("fractaldpos")));
    event InfoLog(address,uint256,uint256);
    function testRead(address miner) public payable{
        consensus.RegisterMiner.value(msg.value)(miner);
        //emit InfoLog(info.OwnerAccount, info.Weight, info.Balance);
    }
}