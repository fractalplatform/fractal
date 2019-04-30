package txpool

import "github.com/fractalplatform/fractal/types"

// TransactionWithPath is the network packet for the transactions.
type TransactionWithPath struct {
	Tx    *types.Transaction
	Bloom *types.Bloom
}
