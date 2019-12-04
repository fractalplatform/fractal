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
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// ErrInvalidSig invalid signature.
var ErrInvalidSig = errors.New("invalid action v, r, s values")

// ActionType type of Action.
type ActionType uint64

const (
	// CallContract represents the call contract action.
	CallContract ActionType = iota
	// CreateContract repesents the create contract action.
	CreateContract
)

const (
	// CreateAccount repesents the create account.
	CreateAccount ActionType = 0x100 + iota
	// UpdateAccount repesents update account.
	UpdateAccount
	// DeleteAccount repesents the delete account action.
	DeleteAccount
	// UpdateAccountAuthor represents the update account author.
	UpdateAccountAuthor
)

const (
	// IncreaseAsset Asset operation
	IncreaseAsset ActionType = 0x200 + iota
	// IssueAsset repesents Issue asset action.
	IssueAsset
	// DestroyAsset destroy asset
	DestroyAsset
	// SetAssetOwner repesents set asset new owner action.
	SetAssetOwner
	// UpdateAsset update asset
	UpdateAsset
	// Transfer repesents transfer asset action.
	Transfer
	UpdateAssetContract
)

const (
	// RegCandidate repesents register candidate action.
	RegCandidate ActionType = 0x300 + iota
	// UpdateCandidate repesents update candidate action.
	UpdateCandidate
	// UnregCandidate repesents unregister candidate action.
	UnregCandidate
	// RefundCandidate repesents unregister candidate action.
	RefundCandidate
	// VoteCandidate repesents voter vote candidate action.
	VoteCandidate

	// UpdateCandidatePubKey repesents update candidate action.
	UpdateCandidatePubKey
)

const (
	// KickedCandidate kicked
	KickedCandidate ActionType = 0x400 + iota
	// ExitTakeOver exit
	ExitTakeOver
	// RemoveKickedCandidate kicked
	RemoveKickedCandidate
)

const (
	// WithdrawFee
	WithdrawFee ActionType = 0x500 + iota
)

type Signature struct {
	ParentIndex uint64
	SignData    []*SignData
}

type SignData struct {
	V     *big.Int
	R     *big.Int
	S     *big.Int
	Index []uint64
}

type FeePayer struct {
	GasPrice *big.Int
	Payer    common.Name
	Sign     *Signature
}

type actionData struct {
	AType    ActionType
	Nonce    uint64
	AssetID  uint64
	From     common.Name
	To       common.Name
	GasLimit uint64
	Amount   *big.Int
	Payload  []byte
	Remark   []byte
	Sign     *Signature

	Extend []rlp.RawValue `rlp:"tail"`
}

// Action represents an entire action in the transaction.
type Action struct {
	data actionData
	// cache
	fp            *FeePayer
	hash          atomic.Value
	extendHash    atomic.Value
	senderPubkeys atomic.Value
	payerPubkeys  atomic.Value
	author        atomic.Value
}

// NewAction initialize transaction's action.
func NewAction(actionType ActionType, from, to common.Name, nonce, assetID, gasLimit uint64, amount *big.Int, payload, remark []byte) *Action {
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
		Sign:     &Signature{0, make([]*SignData, 0)},
		Extend:   make([]rlp.RawValue, 0),
	}
	if amount != nil {
		data.Amount.Set(amount)
	}
	return &Action{data: data}
}

func (a *Action) GetSignIndex(i uint64) []uint64 {
	return a.data.Sign.SignData[i].Index
}

func (a *Action) GetSign() []*SignData {
	return a.data.Sign.SignData
}

func (a *Action) GetSignParent() uint64 {
	return a.data.Sign.ParentIndex
}

func (a *Action) PayerIsExist() bool {
	return a.fp != nil
}

func (a *Action) PayerGasPrice() *big.Int {
	if a.fp != nil {
		return a.fp.GasPrice
	}
	return nil
}

func (a *Action) Payer() common.Name {
	if a.fp != nil {
		return a.fp.Payer
	}
	return common.Name("")
}

func (a *Action) PayerSignature() *Signature {
	if a.fp != nil {
		return a.fp.Sign
	}
	return nil
}

func (a *Action) GetFeePayerSign() []*SignData {
	if a.fp != nil {
		return a.fp.Sign.SignData
	}
	return nil
}

// Check the validity of all fields
func (a *Action) Check(fid uint64, conf *params.ChainConfig) error {
	//check To
	switch a.Type() {
	case CreateContract:
		if a.data.From != a.data.To {
			return fmt.Errorf("Receipt should is %v", a.data.From)
		}
	case CallContract:
	//account
	case CreateAccount:
		fallthrough
	case UpdateAccount:
		fallthrough
	case DeleteAccount:
		fallthrough
	case UpdateAccountAuthor:
		if a.data.To.String() != conf.AccountName {
			return fmt.Errorf("Receipt should is %v", conf.AccountName)
		}
	//asset
	case IncreaseAsset:
		fallthrough
	case IssueAsset:
		fallthrough
	case DestroyAsset:
		fallthrough
	case SetAssetOwner:
		fallthrough
	case UpdateAssetContract:
		fallthrough
	case UpdateAsset:
		if a.data.To.String() != conf.AssetName {
			return fmt.Errorf("Receipt should is %v", conf.AssetName)
		}
	case Transfer:
		//dpos
	case UpdateCandidatePubKey:
		if fid < params.ForkID4 {
			return fmt.Errorf("Receipt undefined")
		}
		fallthrough
	case RegCandidate:
		fallthrough
	case UpdateCandidate:
		fallthrough
	case UnregCandidate:
		fallthrough
	case VoteCandidate:
		fallthrough
	case RefundCandidate:
		fallthrough
	case KickedCandidate:
		fallthrough
	case RemoveKickedCandidate:
		fallthrough
	case ExitTakeOver:
		if a.data.To.String() != conf.DposName {
			return fmt.Errorf("Receipt should is %v", conf.DposName)
		}
		if a.data.AssetID != conf.SysTokenID {
			return fmt.Errorf("Asset id should is %v", conf.SysTokenID)
		}
	default:
		return fmt.Errorf("Receipt undefined")
	}

	//check value
	switch a.Type() {
	case CreateContract:
		fallthrough
	case CallContract:
		fallthrough
	case Transfer:
		fallthrough
	case CreateAccount:
		fallthrough
	case DestroyAsset:
		fallthrough
	case RegCandidate:
		fallthrough
	case UpdateCandidate:
		return nil
	default:
	}
	if a.Value().Cmp(big.NewInt(0)) != 0 {
		return fmt.Errorf("Value should is zero")
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
func (a *Action) Sender() common.Name { return a.data.From }

// Recipient returns action's Recipient.
func (a *Action) Recipient() common.Name { return a.data.To }

// Data returns action's payload.
func (a *Action) Data() []byte { return common.CopyBytes(a.data.Payload) }

// Remark returns action's remark.
func (a *Action) Remark() []byte { return common.CopyBytes(a.data.Remark) }

// Gas returns action's Gas.
func (a *Action) Gas() uint64 { return a.data.GasLimit }

// Value returns action's Value.
func (a *Action) Value() *big.Int { return new(big.Int).Set(a.data.Amount) }

func (a *Action) Extend() []rlp.RawValue { return a.data.Extend }

// IgnoreExtend returns ignore extend
func (a *Action) IgnoreExtend() []interface{} {
	return []interface{}{
		a.data.AType,
		a.data.Nonce,
		a.data.AssetID,
		a.data.From,
		a.data.To,
		a.data.GasLimit,
		a.data.Amount,
		a.data.Payload,
		a.data.Remark,
		a.data.Sign,
	}
}

// EncodeRLP implements rlp.Encoder
func (a *Action) EncodeRLP(w io.Writer) error {
	if a.fp != nil {
		value, err := rlp.EncodeToBytes(a.fp)
		if err != nil {
			return err
		}
		a.data.Extend = []rlp.RawValue{value}
	}

	return rlp.Encode(w, &a.data)
}

// DecodeRLP implements rlp.Decoder
func (a *Action) DecodeRLP(s *rlp.Stream) error {
	if err := s.Decode(&a.data); err != nil {
		return err
	}

	if len(a.data.Extend) != 0 {
		a.fp = new(FeePayer)
		return rlp.DecodeBytes(a.data.Extend[0], a.fp)

	}

	return nil
}

// ChainID returns which chain id this action was signed for (if at all)
func (a *Action) ChainID() *big.Int {
	return deriveChainID(a.data.Sign.SignData[0].V)
}

// Hash hashes the RLP encoding of action.
func (a *Action) Hash() common.Hash {
	if hash := a.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := RlpHash(a.IgnoreExtend())
	a.hash.Store(v)
	return v
}

// ExtendHash hashes the RLP encoding of action.
func (a *Action) ExtendHash() common.Hash {
	if hash := a.extendHash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := RlpHash(a)
	a.extendHash.Store(v)
	return v
}

// WithSignature returns a new transaction with the given signature.
func (a *Action) WithSignature(signer Signer, sig []byte, index []uint64) error {
	r, s, v, err := signer.SignatureValues(sig)
	if err != nil {
		return err
	}
	a.data.Sign.SignData = append(a.data.Sign.SignData, &SignData{R: r, S: s, V: v, Index: index})
	return nil
}

// WithParentIndex returns a new transaction with the given signature.
func (a *Action) WithParentIndex(parentIndex uint64) {
	a.data.Sign.ParentIndex = parentIndex
}

func (f *FeePayer) WithSignature(signer Signer, sig []byte, index []uint64) error {
	r, s, v, err := signer.SignatureValues(sig)
	if err != nil {
		return err
	}
	f.Sign.SignData = append(f.Sign.SignData, &SignData{R: r, S: s, V: v, Index: index})
	return nil
}

func (f *FeePayer) WithParentIndex(parentIndex uint64) {
	f.Sign.ParentIndex = parentIndex
}

func (f *FeePayer) GetSignParent() uint64 {
	return f.Sign.ParentIndex
}

func (f *FeePayer) GetSignIndex(i uint64) []uint64 {
	return f.Sign.SignData[i].Index
}

// RPCAction represents a action that will serialize to the RPC representation of a action.
type RPCAction struct {
	Type       uint64        `json:"type"`
	Nonce      uint64        `json:"nonce"`
	From       common.Name   `json:"from"`
	To         common.Name   `json:"to"`
	AssetID    uint64        `json:"assetID"`
	GasLimit   uint64        `json:"gas"`
	Amount     *big.Int      `json:"value"`
	Remark     hexutil.Bytes `json:"remark"`
	Payload    hexutil.Bytes `json:"payload"`
	Hash       common.Hash   `json:"actionHash"`
	ActionIdex uint64        `json:"actionIndex"`
}

func (a *RPCAction) SetHash(hash common.Hash) {
	a.Hash = hash
}

// NewRPCAction returns a action that will serialize to the RPC.
func (a *Action) NewRPCAction(index uint64) *RPCAction {
	return &RPCAction{
		Type:       uint64(a.Type()),
		Nonce:      a.Nonce(),
		From:       a.Sender(),
		To:         a.Recipient(),
		AssetID:    a.AssetID(),
		GasLimit:   a.Gas(),
		Amount:     a.Value(),
		Remark:     hexutil.Bytes(a.Remark()),
		Payload:    hexutil.Bytes(a.Data()),
		Hash:       a.Hash(),
		ActionIdex: index,
	}
}

type RPCActionWithPayer struct {
	Type          uint64        `json:"type"`
	Nonce         uint64        `json:"nonce"`
	From          common.Name   `json:"from"`
	To            common.Name   `json:"to"`
	AssetID       uint64        `json:"assetID"`
	GasLimit      uint64        `json:"gas"`
	Amount        *big.Int      `json:"value"`
	Remark        hexutil.Bytes `json:"remark"`
	Payload       hexutil.Bytes `json:"payload"`
	Hash          common.Hash   `json:"actionHash"`
	ActionIdex    uint64        `json:"actionIndex"`
	Payer         common.Name   `json:"payer"`
	PayerGasPrice *big.Int      `json:"payerGasPrice"`
}

func (a *RPCActionWithPayer) SetHash(hash common.Hash) {
	a.Hash = hash
}

func (a *Action) NewRPCActionWithPayer(index uint64) *RPCActionWithPayer {
	var payer common.Name
	var price *big.Int
	if a.fp != nil {
		payer = a.fp.Payer
		price = a.fp.GasPrice
	}
	return &RPCActionWithPayer{
		Type:          uint64(a.Type()),
		Nonce:         a.Nonce(),
		From:          a.Sender(),
		To:            a.Recipient(),
		AssetID:       a.AssetID(),
		GasLimit:      a.Gas(),
		Amount:        a.Value(),
		Remark:        hexutil.Bytes(a.Remark()),
		Payload:       hexutil.Bytes(a.Data()),
		Hash:          a.Hash(),
		ActionIdex:    index,
		Payer:         payer,
		PayerGasPrice: price,
	}
}

// deriveChainID derives the chain id from the given v parameter
func deriveChainID(v *big.Int) *big.Int {
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}
