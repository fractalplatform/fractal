pragma solidity >=0.4.0;

contract Plugin {
    function Call(address plugin) internal {
        plugin.call(msg.data);
        assembly {
            let rsize := returndatasize
            let roff := mload(0x40)
            returndatacopy(roff, 0, rsize)
            return(roff, rsize)
        }
    }
}