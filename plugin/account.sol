pragma solidity >=0.4.0;
pragma experimental ABIEncoderV2;

contract AccountAPI {
    function CreateAccount(string name, string pubKey, string desc) external;
    function ChangePubKey(string pubKey) external;
    function GetBalance(string to,uint64 assetid) external returns(uint256);
    function Transfer(string to, uint64 assetid, uint256 value) external;
}

contract TestAccount {
    AccountAPI account = AccountAPI(address(bytes20("fractalaccount")));

    function testCreateAccount(string name, string pubKey, string desc) public {
        account.CreateAccount(name, pubKey, desc);
    }

    function testChangePubKey(string pubKey) public {
        account.ChangePubKey(pubKey);
    }

    function testGetBalance(string toname, uint64 assetid) public {
        account.GetBalance(toname,assetid);
    }

    function testTransfer(string to, uint64 assetid, uint256 value) public {
        account.Transfer(to, assetid, value);
    }
}