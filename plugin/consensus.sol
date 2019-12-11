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
    function GetMinerInfo(address miner) public returns(MinerInfo memory){
        if (msg.sender == address(this))
            Call();
        else
            return this.GetMinerInfo(miner);
    }

    function UnregisterMiner() public {
        if (msg.sender == address(this))
            Call();
        else
            return this.UnregisterMiner();
    }

    function RegisterMiner(address miner) public {
        if (msg.sender == address(this))
            Call();
        else
            return this.RegisterMiner(miner);
    }
    event InfoLog(MinerInfo);
    function testRead(address miner) public {
        MinerInfo memory info = GetMinerInfo(miner);
        emit InfoLog(info);
    }
}