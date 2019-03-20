package sdk

// DposInfo dpos info
func (api *API) DposInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_info")
	return info, err
}

// DposIrreversible dpos irreversible info
func (api *API) DposIrreversible() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_irreversible")
	return info, err
}

// DposEpcho dpos state info by height
func (api *API) DposEpcho(height uint64) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_epcho", height)
	return info, err
}

// DposLatestEpcho dpos state info by height
func (api *API) DposLatestEpcho() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_latestEpcho")
	return info, err
}

// DposValidateEpcho dpos state info
func (api *API) DposValidateEpcho() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_validateEpcho")
	return info, err
}

// DposCadidates dpos cadidate info
func (api *API) DposCadidates() ([]map[string]interface{}, error) {
	info := []map[string]interface{}{}
	err := api.client.Call(&info, "dpos_cadidates")
	return info, err
}

// DposAccount dpos account info
func (api *API) DposAccount(name string) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_account", name)
	return info, err
}
