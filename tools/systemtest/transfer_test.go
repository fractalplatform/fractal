package main

import (
	"fmt"
	"github.com/fractalplatform/systemtest/rpc"
	"testing"
)

func TestGetTxpoolInfo(t *testing.T) {
	pengingNum, queueNum := rpc.GetTxpoolStatus()
	fmt.Printf("pending = %d, queueNum = %d", pengingNum, queueNum)
}
