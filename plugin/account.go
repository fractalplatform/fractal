package plugin

type NativeAccount struct {
}

func (c *NativeAccount) Run(method string, params ...interface{}) ([]byte, error) {
	return nil, nil
}

func (c *NativeAccount) GetNonce(params ...interface{}) uint64 {
	return 1
}
