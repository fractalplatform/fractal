package common

// MinerStart start
func (api *API) MinerStart() (bool, error) {
	ret := false
	err := api.client.Call(&ret, "miner_start")
	return ret, err
}

// MinerStop stop
func (api *API) MinerStop() (bool, error) {
	ret := false
	err := api.client.Call(&ret, "miner_stop")
	return ret, err
}

// MinerMining mining
func (api *API) MinerMining() (bool, error) {
	ret := false
	err := api.client.Call(&ret, "miner_mining")
	return ret, err
}

// MinerSetExtra extra
func (api *API) MinerSetExtra(extra []byte) (bool, error) {
	var verr error
	err := api.client.Call(&verr, "miner_setExtra", extra)
	return verr == nil, err
}

// MinerSetCoinbase coinbase
func (api *API) MinerSetCoinbase(name string, privKey string) (bool, error) {
	var verr error
	err := api.client.Call(&verr, "miner_setCoinbase", name, privKey)
	return verr == nil, err
}
