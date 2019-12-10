package envelope

import (
	"math/big"
)

type BasicTx struct {
	Typ        Type
	GasAssetID uint64
	GasPrice   *big.Int
	GasLimit   uint64
	Nonce      uint64
	From       string
	To         string
}

func (b *BasicTx) Type() Type            { return b.Typ }
func (b *BasicTx) GetGasAssetID() uint64 { return b.GasAssetID }
func (b *BasicTx) GetGasLimit() uint64   { return b.GasLimit }
func (b *BasicTx) GetGasPrice() *big.Int { return b.GasPrice }
func (b *BasicTx) GetNonce() uint64      { return b.Nonce }
func (b *BasicTx) Sender() string        { return b.From }
func (b *BasicTx) Recipient() string     { return b.To }
