package sdk

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
	ret := true
	err := api.client.Call(&ret, "miner_setExtra", extra)
	return ret, err
}

// MinerSetCoinbase coinbase
func (api *API) MinerSetCoinbase(name string, privKeys []string) (bool, error) {
	ret := true
	err := api.client.Call(&ret, "miner_setCoinbase", name, privKeys)
	return ret, err
}
