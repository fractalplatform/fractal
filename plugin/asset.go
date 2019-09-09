package plugin

type NativeAsset struct {
}

func (contract *NativeAsset) Run(method string, params ...interface{}) ([]byte, error) {
	return nil, nil
}
