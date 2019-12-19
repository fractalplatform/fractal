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
    function RegisterMiner(address signer) external payable;
}

contract TestConsensus {
    ConsensusAPI constant consensus = ConsensusAPI(address(bytes20("fractaldpos")));
    function testReadInfo() public returns(ConsensusAPI.MinerInfo memory){
        ConsensusAPI.MinerInfo memory info = consensus.GetMinerInfo(address(this));
        return info;
    }

    function testRegister(uint256 amount) public payable {
        consensus.RegisterMiner.value(amount)(msg.sender);
    }

    function testUnregister() public {
        consensus.UnregisterMiner();
    }
}