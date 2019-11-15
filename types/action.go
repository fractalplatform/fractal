// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// ActionType type of Action.
type ActionType uint64

const (
	// CallContract represents the call contract action.
	CallContract ActionType = iota
	// CreateContract represents the create contract action.
	CreateContract
	// Transfer represents transfer asset action.
	Transfer
)

type actionData struct {
	AType     ActionType
	Nonce     uint64
	AssetID   uint64
	From      string
	To        string
	GasLimit  uint64
	Amount    *big.Int
	Payload   []byte
	Remark    []byte
	Signature []byte
}

// Action represents an entire action in the transaction.
type Action struct {
	data actionData
	// cache
	hash          atomic.Value
	senderPubkeys atomic.Value
	author        atomic.Value
}

// NewAction initialize transaction's action.
func NewAction(actionType ActionType, from, to string, nonce, assetID,
	gasLimit uint64, amount *big.Int, payload, remark []byte) *Action {
	if len(payload) > 0 {
		payload = common.CopyBytes(payload)
	}
	data := actionData{
		AType:    actionType,
		Nonce:    nonce,
		AssetID:  assetID,
		From:     from,
		To:       to,
		GasLimit: gasLimit,
		Amount:   new(big.Int),
		Payload:  payload,
		Remark:   remark,
	}
	if amount != nil {
		data.Amount.Set(amount)
	}
	return &Action{data: data}
}

// GetSign returns action signature
func (a *Action) GetSign() []byte {
	return a.data.Signature
}

// Check the validity of all fields
func (a *Action) Check() error {
	//check To
	switch a.Type() {
	case CreateContract:
		if a.data.From != a.data.To {
			return fmt.Errorf("Receipt should is %v", a.data.From)
		}
	}
	return nil
}

// Type returns action's type.
func (a *Action) Type() ActionType { return a.data.AType }

// Nonce returns action's nonce.
func (a *Action) Nonce() uint64 { return a.data.Nonce }

// AssetID returns action's assetID.
func (a *Action) AssetID() uint64 { return a.data.AssetID }

// Sender returns action's Sender.
func (a *Action) Sender() string { return a.data.From }

// Recipient returns action's Recipient.
func (a *Action) Recipient() string { return a.data.To }

// Data returns action's payload.
func (a *Action) Data() []byte { return common.CopyBytes(a.data.Payload) }

// Remark returns action's remark.
func (a *Action) Remark() []byte { return common.CopyBytes(a.data.Remark) }

// Gas returns action's Gas.
func (a *Action) Gas() uint64 { return a.data.GasLimit }

// Value returns action's Value.
func (a *Action) Value() *big.Int { return new(big.Int).Set(a.data.Amount) }

// EncodeRLP implements rlp.Encoder
func (a *Action) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &a.data)
}

// DecodeRLP implements rlp.Decoder
func (a *Action) DecodeRLP(s *rlp.Stream) error {
	return s.Decode(&a.data)
}

// Hash hashes the RLP encoding of action.
func (a *Action) Hash() common.Hash {
	if hash := a.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := RlpHash(a)
	a.hash.Store(v)
	return v
}

// WithSignature returns a new transaction with the given signature.
func (a *Action) WithSignature(sig []byte) {
	a.data.Signature = sig
}

// RPCAction represents a action that will serialize to the RPC representation of a action.
type RPCAction struct {
	Type        uint64        `json:"type"`
	Nonce       uint64        `json:"nonce"`
	From        string        `json:"from"`
	To          string        `json:"to"`
	AssetID     uint64        `json:"assetID"`
	GasLimit    uint64        `json:"gas"`
	Amount      *big.Int      `json:"value"`
	Remark      hexutil.Bytes `json:"remark"`
	Payload     hexutil.Bytes `json:"payload"`
	Hash        common.Hash   `json:"hash"`
	ActionIndex uint64        `json:"index"`
}

// NewRPCAction returns a action that will serialize to the RPC.
func (a *Action) NewRPCAction(index uint64) *RPCAction {
	return &RPCAction{
		Type:        uint64(a.Type()),
		Nonce:       a.Nonce(),
		From:        a.Sender(),
		To:          a.Recipient(),
		AssetID:     a.AssetID(),
		GasLimit:    a.Gas(),
		Amount:      a.Value(),
		Remark:      hexutil.Bytes(a.Remark()),
		Payload:     hexutil.Bytes(a.Data()),
		Hash:        a.Hash(),
		ActionIndex: index,
	}
}
