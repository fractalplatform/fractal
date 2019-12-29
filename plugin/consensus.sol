pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

contract ConsensusAPI {
    struct MinerInfo {
        address OwnerAccount;
        address SignAccount;
        uint256 RegisterNumber;
        uint256 Weight;
        uint256 Balance;
        uint256 Epoch;
    }
    function GetMinerInfo(string miner) external returns(MinerInfo memory);
    function UnregisterMiner() external;
    function RegisterMiner(string signer) external payable;
}

contract TestConsensus {
    ConsensusAPI constant consensus = ConsensusAPI(address(bytes20("fractaldpos")));
    function testReadInfo(string miner) public returns(ConsensusAPI.MinerInfo memory){
        ConsensusAPI.MinerInfo memory info = consensus.GetMinerInfo(miner);
        return info;
    }

    function testRegister(uint256 amount, string signer) public payable {
        consensus.RegisterMiner.value(amount)(signer);
    }

    function testUnregister() public {
        consensus.UnregisterMiner();
    }
}