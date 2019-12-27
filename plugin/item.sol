pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

interface ItemAPI {
    struct MinerInfo {
        address OwnerAccount;
        address SignAccount;
        uint256 RegisterNumber;
        uint256 Weight;
        uint256 Balance;
        uint256 Epoch;
    }
    function IssueWorld(address owner, string name, string description) external;
    function IssueItemType(uint64 worldID, string name, bool merge, uint64 upperLimit, string description, uint64[] attrPermission, string[] attrName, string[] attrDes) external;
    function IncreaseItem(uint64 worldID, uint64 itemTypeID, address owner, string description, uint64[] attrPermission, string[] attrName, string[] attrDes) external;
    function DestroyItem(uint64 worldID, uint64 itemTypeID, uint64 itemID) external;
    function IncreaseItems(uint64 worldID, uint64 itemTypeID, address to, uint64 amount) external;
    function DestroyItems(uint64 worldID, uint64 itemTypeID, uint64 amount) external;
    function TransferItem(address to, uint64[] worldID, uint64[] itemTypeID, uint64[] itemID, uint64[] amount) external;
    function AddItemTypeAttributes(uint64 worldID, uint64 itemTypeID, uint64[] attrPermission, string[] attrName, string[] attrDes) external;
    function DelItemTypeAttributes(uint64 worldID, uint64 itemTypeID, string[] attrName) external;
    function ModifyItemTypeAttributes(uint64 worldID, uint64 itemTypeID, uint64[] attrPermission, string[] attrName, string[] attrDes) external;
    function AddItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, uint64[] attrPermission, string[] attrName, string[] attrDes) external;
    function DelItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, string[] attrName) external;
    function ModifyItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, uint64[] attrPermission, string[] attrName, string[] attrDes) external;
}

contract TestConsensus {
    ItemAPI constant item = ItemAPI(address(bytes20("fractalitem")));
    // function testReadInfo() public returns(ConsensusAPI.MinerInfo memory){
    //     ConsensusAPI.MinerInfo memory info = consensus.GetMinerInfo(address(this));
    //     return info;
    // }

    function testIssueWorld(address owner, string name, string description) public {
        item.IssueWorld(owner, name, description);
    }

    function testIssueItemType(uint64 worldID, string name, bool merge, uint64 upperLimit, string description, uint64[] attrPermission, string[] attrName, string[] attrDes) public {
        item.IssueItemType(worldID, name, merge, upperLimit, description, attrPermission, attrName, attrDes);
    }

    function testIncreaseItem(uint64 worldID, uint64 itemTypeID, address owner, string description, uint64[] attrPermission, string[] attrName, string[] attrDes) public {
        item.IncreaseItem(worldID, itemTypeID, owner, description, attrPermission, attrName, attrDes);
    }

    function testDestroyItem(uint64 worldID, uint64 itemTypeID, uint64 itemID) public {
        item.DestroyItem(worldID, itemTypeID, itemID);
    }

    function testIncreaseItems(uint64 worldID, uint64 itemTypeID, address to, uint64 amount) public {
        item.IncreaseItems(worldID, itemTypeID, to, amount);
    }

    function testDestroyItems(uint64 worldID, uint64 itemTypeID, uint64 amount) public {
        item.DestroyItems(worldID, itemTypeID, amount);
    }

    function testAddItemTypeAttributes(uint64 worldID, uint64 itemTypeID, uint64[] attrPermission, string[] attrName, string[] attrDes) public {
        item.AddItemTypeAttributes(worldID, itemTypeID, attrPermission, attrName, attrDes);
    }
    
    function testDelItemTypeAttributes(uint64 worldID, uint64 itemTypeID, string[] attrName) public {
        item.DelItemTypeAttributes(worldID, itemTypeID, attrName);
    }

    function testModifyItemTypeAttributes(uint64 worldID, uint64 itemTypeID, uint64[] attrPermission, string[] attrName, string[] attrDes) public {
        item.ModifyItemTypeAttributes(worldID, itemTypeID, attrPermission, attrName, attrDes);
    }

    function testAddItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, uint64[] attrPermission, string[] attrName, string[] attrDes) public {
        item.AddItemAttributes(worldID, itemTypeID, itemID, attrPermission, attrName, attrDes);
    }

    function testDelItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, string[] attrName) public {
        item.DelItemAttributes(worldID, itemTypeID, itemID, attrName);
    }

    function testModifyItemAttributes(uint64 worldID, uint64 itemTypeID, uint64 itemID, uint64[] attrPermission, string[] attrName, string[] attrDes) public {
        item.ModifyItemAttributes(worldID, itemTypeID, itemID, attrPermission, attrName, attrDes);
    }
}