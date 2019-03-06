package main

import (
	tc "github.com/fractalplatform/fractal/test/common"
	"github.com/fractalplatform/fractal/rpc"
	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)
}

func main() {
	for i := 0; i < 100; i++ {
		result, err := tc.GetBlockAndResult(rpc.BlockNumber(i))
		if err != nil {
			jww.ERROR.Println("get block and result failed", err)
		}
		jww.INFO.Println(result)
	}
}
