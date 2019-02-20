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
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/crypto"
)

var (
	//ErrInvalidchainID invalid chain id for signer
	ErrInvalidchainID = errors.New("invalid chain id for signer")
	//ErrSigUnprotected signature is considered unprotected
	ErrSigUnprotected = errors.New("signature is considered unprotected")
)

// sigCache is used to cache the derived sender and contains the signer used to derive it.
type sigCache struct {
	signer Signer
	pubKey []byte
}

// MakeSigner returns a Signer based on the given chainID .
func MakeSigner(chainID *big.Int) Signer {
	return NewSigner(chainID)
}

// SignAction signs the action using the given signer and private key
func SignAction(a *Action, tx *Transaction, s Signer, prv *ecdsa.PrivateKey) error {
	h := s.Hash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return err
	}
	return a.WithSignature(s, sig)
}

// Recover returns the pubkey derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
func Recover(signer Signer, a *Action, tx *Transaction) (common.PubKey, error) {
	if sc := a.sender.Load(); sc != nil {
		sigCache := sc.(sigCache)
		if sigCache.signer.Equal(signer) {
			pk := new(common.PubKey)
			pk.SetBytes(sigCache.pubKey)
			return *pk, nil
		}
	}

	pubKey, err := signer.PubKey(a, tx)
	if err != nil {
		return common.PubKey{}, err
	}
	a.sender.Store(sigCache{signer: signer, pubKey: pubKey})
	return common.BytesToPubKey(pubKey), nil
}

// Signer implements Signer .
type Signer struct {
	chainID, chainIDMul *big.Int
}

// NewSigner initialize signer
func NewSigner(chainID *big.Int) Signer {
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Signer{
		chainID:    chainID,
		chainIDMul: new(big.Int).Mul(chainID, big.NewInt(2)),
	}
}

// Equal judging the same chainID
func (s Signer) Equal(s2 Signer) bool {
	return s2.chainID.Cmp(s.chainID) == 0
}

var big8 = big.NewInt(8)

// PubKey return Action sender
func (s Signer) PubKey(a *Action, tx *Transaction) ([]byte, error) {
	if a.ChainID().Cmp(s.chainID) != 0 {
		return nil, ErrInvalidchainID
	}
	V := new(big.Int).Sub(a.data.V, s.chainIDMul)
	V.Sub(V, big8)
	return recoverPlain(s.Hash(tx), a.data.R, a.data.S, V, true)
}

// SignatureValues returns a new transaction with the given signature. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (s Signer) SignatureValues(sig []byte) (R, S, V *big.Int, err error) {
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

// Hash returns the hash to be signed by the sender.
func (s Signer) Hash(tx *Transaction) common.Hash {
	actionHashs := make([]common.Hash, len(tx.GetActions()))
	for _, a := range tx.GetActions() {
		hash := rlpHash([]interface{}{
			a.data.From,
			a.data.AType,
			a.data.Nonce,
			a.data.To,
			a.data.GasLimit,
			a.data.Amount,
			a.data.Payload,
			s.chainID, uint(0), uint(0),
		})
		actionHashs = append(actionHashs, hash)
	}

	return rlpHash([]interface{}{
		common.MerkleRoot(actionHashs),
		tx.gasAssetID,
		tx.gasPrice,
	})
}

func recoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) ([]byte, error) {
	if Vb.BitLen() > 8 {
		return nil, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
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
