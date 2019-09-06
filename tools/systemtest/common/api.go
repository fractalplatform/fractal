package common

import (
	"fmt"

	"github.com/fractalplatform/fractal/rpc"
)

// API rpc api
type API struct {
	rpchost string
	client  *rpc.Client
}

// NewAPI create api interface
func NewAPI(rpchost string) *API {
	client, err := rpc.DialHTTP(rpchost)
	if err != nil {
		panic(fmt.Sprintf("dial http %v err %v", rpchost, err))
	}
	api := &API{}
	api.rpchost = rpchost
	api.client = client
	return api
}
