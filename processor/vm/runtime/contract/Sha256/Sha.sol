pragma solidity ^0.4.24;

contract Sha {
    constructor() public payable {
    }
    
    function sha(string name) public {
        sha256(name);
    }
}
