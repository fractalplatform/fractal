package envelope

import (
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/common"
)

// Type type of Action.
type Type uint64

const (
	Unknown Type = iota
	// CallContract represents the call contract .
	CallContract
	// CreateContract represents the create contract .
	CreateContract

	Plugin
)

type Envelope interface {
	Type() Type
	GetGasAssetID() uint64
	GetGasLimit() uint64
	GetGasPrice() *big.Int
	GetNonce() uint64
	Sender() string
	Recipient() string
	Cost() *big.Int
	Hash() common.Hash
	GetSign() []byte
	//sign
	SignHash(chainID *big.Int) common.Hash

	// rpc
	NewRPCTransaction(blockHash common.Hash, blockNumber uint64, index uint64) interface{}
}

func New(txType Type) (Envelope, error) {
	switch txType {
	case CallContract:
		fallthrough
	case CreateContract:
		return &ContractTx{}, nil
	case Plugin:
		return &PluginTx{}, nil
	}
	return nil, fmt.Errorf("unknown payload type: %d", txType)
}
