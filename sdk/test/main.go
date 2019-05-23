package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/fractalplatform/fractal/sdk"
)

func main() {
	api := sdk.NewAPI("")
	txs := []*TTX{}
	total := 0
	failed := 0
	for index, tx := range txs {
		total++
		err := runTx(api, tx)
		if err != nil {
			failed++
		}
		fmt.Println(fmt.Sprintf("%5d %v", index, err))
	}
	fmt.Println("total", total, "failed", failed)
}

func runTx(api *sdk.API, tx *TTX) error {
	switch strings.ToLower(tx.Type) {
	case "createaccount":
	default:
	}
	return nil
}

// TTX
type TTX struct {
	Type    string      `json:"type,omitempty"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
	AssetID uint64      `json:"id,omitempty"`
	Value   *big.Int    `json:"value,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
	Childs  []*TTX      `json:"childs,omitempty"`
	Succeed bool        `json:"succeed,omitempty"`
	Contain string      `json:"contain,omitempty"`
}
