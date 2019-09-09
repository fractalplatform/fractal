package rpc

func GetPeerCount(nodeIp string, nodePort int64) (int, error) {
	peerCount := int(0)
	err := ClientCallWithAddr(nodeIp, nodePort, "p2p_peerCount", &peerCount)

	return peerCount, err
}
