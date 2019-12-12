pragma solidity >=0.4.0;

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