pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

contract Plugin {
    function Call() internal {
        address(bytes20("fractaldpos")).call(msg.data);
        assembly {
            let rsize := returndatasize
            let roff := mload(0x40)
            returndatacopy(roff, 0, rsize)
            return(roff, rsize)
        }
    }
}

contract ConsensusAPI is Plugin {
    struct MinerInfo {
        address OwnerAccount;
        address SignAccount;
        uint256 RegisterNumber;
        uint256 Weight;
        uint256 Balance;
        uint256 Epoch;
    }
    function GetMinerInfo(address miner) external returns(MinerInfo memory){
        Plugin.Call();
    }

    function UnregisterMiner() public {
        Plugin.Call();
    }
    function RegisterMiner(address miner) external payable {
        Plugin.Call();
    }

    event InfoLog(address);
    function testRead(address miner) public {
        MinerInfo memory info = this.GetMinerInfo(miner);
        emit InfoLog(info.OwnerAccount);
    }
}