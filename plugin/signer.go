// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
	"github.com/fractalplatform/fractal/types"
	"math/big"
)

var (
	//ErrInvalidchainID invalid chain id for signer
	ErrInvalidChainID = errors.New("invalid chain id for signer")
	//ErrSigUnprotected signature is considered unprotected
	ErrSignUnprotected = errors.New("signature is considered unprotected")
	//ErrSignEmpty signature is considered unprotected
	ErrSignEmpty = errors.New("signature is nil")
	//ErrInvalidSig invalid signature.
	ErrInvalidSig = errors.New("invalid action v, r, s values")
)

type SignData struct {
	V *big.Int
	R *big.Int
	S *big.Int
}

type Signer struct {
	chainId    *big.Int
	chainIdMul *big.Int
}

func NewSigner(chainId *big.Int) (ISigner, error) {
	if chainId == nil {
		chainId = new(big.Int)
	}
	return &Signer{
		chainId:    chainId,
		chainIdMul: new(big.Int).Mul(chainId, big.NewInt(2)),
	}, nil
}

func (s *Signer) Sign(data interface{}, prv *ecdsa.PrivateKey) ([]byte, error) {
	var h common.Hash
	switch d := data.(type) {
	case []byte:
		h = types.RlpHash(d)
	case types.Action:
		actionData := d.Data()
		h = types.RlpHash(actionData)
	case types.Transaction:
		h = types.RlpHash(d)
	default:
		return nil, errors.New("signer: unknown data type")

	}
	signData, err := crypto.Sign(h[:], prv)
	return signData, err
}

func (s *Signer) Hash() {

}

func getChainID(action *types.Action) *big.Int {
	signData := action.GetSign()
	v := big.NewInt(int64(signData[64]))
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}

func (s *Signer) Recover(action *types.Action, tx *types.Transaction) ([]byte, error) {
	signData := action.GetSign()
	R, S, V, err := s.SignatureValues(signData)
	if err != nil {
		return nil, err
	}
	chainIdMul := new(big.Int).Sub(V, big.NewInt(35))
	chainID := chainIdMul.Div(chainIdMul, big.NewInt(2))
	if chainID.Cmp(s.chainId) != 0 {
		return nil, ErrInvalidChainID
	}
	V = new(big.Int).Sub(V, chainIdMul)
	V.Sub(V, big.NewInt(8))
	data, err := recoverPlain(types.RlpHash(tx), R, S, V)
	//pubKey := common.BytesToPubKey(data)
	return data, nil
}

func recoverPlain(sighash common.Hash, R, S, Vb *big.Int) ([]byte, error) {
	if Vb.BitLen() > 8 {
		return nil, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S) {
		return nil, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the snature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return nil, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return nil, errors.New("invalid public key")
	}
	return pub, nil
}

// SignatureValues returns a new transaction with the given signature. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (s *Signer) SignatureValues(sig []byte) (R, S, V *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	R = new(big.Int).SetBytes(sig[:32])
	S = new(big.Int).SetBytes(sig[32:64])
	V = new(big.Int).SetBytes([]byte{sig[64] + 27})

	return R, S, V, nil
}
