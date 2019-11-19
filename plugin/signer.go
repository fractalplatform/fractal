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
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
)

var (
	//ErrInvalidChainID invalid chain id for signer
	ErrInvalidChainID = errors.New("invalid chain id for signer")
	//ErrInvalidSig invalid signature.
	ErrInvalidSig = errors.New("invalid action v, r, s values")
)

type SignData struct {
	V *big.Int
	R *big.Int
	S *big.Int
}

type Signer struct {
	chainID    *big.Int
	chainIDMul *big.Int
}

func NewSigner(chainID *big.Int) (ISigner, error) {
	if chainID == nil {
		chainID = new(big.Int)
	}
	return &Signer{
		chainID:    chainID,
		chainIDMul: new(big.Int).Mul(chainID, big.NewInt(2)),
	}, nil
}

func (s *Signer) Sign(signHash common.Hash, prv *ecdsa.PrivateKey) ([]byte, error) {
	sigBytes, err := crypto.Sign(signHash[:], prv)
	if err != nil {
		return nil, err
	}

	R, S, V, err := s.signatureValues(sigBytes)
	if err != nil {
		return nil, err
	}

	rb, sb := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(rb):32], rb)
	copy(sig[64-len(sb):64], sb)
	sig[64] = byte(V.Uint64())

	return sig, nil
}

var big8 = big.NewInt(8)

func (s *Signer) Recover(signature []byte, signHash common.Hash) ([]byte, error) {
	R := new(big.Int).SetBytes(signature[:32])
	S := new(big.Int).SetBytes(signature[32:64])
	V := new(big.Int).SetBytes([]byte{signature[64]})

	chainID := deriveChainID(V)
	if chainID.Cmp(s.chainID) != 0 {
		return nil, ErrInvalidChainID
	}

	V = new(big.Int).Sub(V, s.chainIDMul)
	V.Sub(V, big8)

	return recoverPlain(signHash, R, S, V)
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
func (s *Signer) signatureValues(sig []byte) (R, S, V *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	R = new(big.Int).SetBytes(sig[:32])
	S = new(big.Int).SetBytes(sig[32:64])
	V = new(big.Int).SetBytes([]byte{sig[64] + 27})

	if s.chainID.Sign() != 0 {
		V = big.NewInt(int64(sig[64] + 35))
		V.Add(V, s.chainIDMul)
	}

	return R, S, V, nil
}

// deriveChainID derives the chain id from the given v parameter
func deriveChainID(v *big.Int) *big.Int {
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}
