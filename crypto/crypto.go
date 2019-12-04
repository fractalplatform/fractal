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

package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

var errInvalidPubkey = errors.New("invalid secp256k1 public key")

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := common.Get256()
	defer common.Put256(d)
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// Keccak256Hash calculates and returns the Keccak256 hash of the input data,
// converting it to an internal Hash data structure.
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := common.Get256()
	defer common.Put256(d)
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Keccak512 calculates and returns the Keccak512 hash of the input data.
func Keccak512(data ...[]byte) []byte {
	d := common.Get512()
	defer common.Put512(d)
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// CreateAddress creates an  address given args
func CreateAddress(args ...interface{}) common.Address {
	data, err := rlp.EncodeToBytes(args)
	if err != nil {
		panic(err)
	}
	return common.BytesToAddress(Keccak256(data)[12:])
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

// ToECDSAUnsafe blindly converts a binary blob to a private key. It should almost
// never be used unless you are sure the input is valid and want to avoid hitting
// errors due to bad origin encoding (0 prefixes cut off).
func ToECDSAUnsafe(d []byte) *ecdsa.PrivateKey {
	priv, _ := toECDSA(d, false)
	return priv
}

// toECDSA creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

// FromECDSA exports a private key into a binary dump.
func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

// UnmarshalPubkey converts bytes to a secp256k1 public key.
func UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(S256(), pub)
	if x == nil {
		return nil, errInvalidPubkey
	}
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	return ToECDSA(b)
}

// LoadECDSA loads a secp256k1 private key from the given file.
func LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}
	return ToECDSA(key)
}

// SaveECDSA saves a secp256k1 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveECDSA(file string, key *ecdsa.PrivateKey) error {
	k := hex.EncodeToString(FromECDSA(key))
	return ioutil.WriteFile(file, []byte(k), 0600)
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(S256(), rand.Reader)
}

// ValidateSignatureValues verifies whether the signature values are valid with
// the given chain rules. The v value is assumed to be either 0 or 1.
func ValidateSignatureValues(v byte, r, s *big.Int) bool {
	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if s.Cmp(secp256k1halfN) > 0 {
		return false
	}
	// Frontier: allow s to be in full N range
	return r.Cmp(secp256k1N) < 0 && s.Cmp(secp256k1N) < 0 && (v == 0 || v == 1)
}

func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&p)
	return common.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

func PubkeyAdd(p1 ecdsa.PublicKey, p2 ecdsa.PublicKey) ecdsa.PublicKey {
	retPub := p1
	retPub.X, retPub.Y = p1.Add(p1.X, p1.Y, p2.X, p2.Y)
	return retPub
}

func PrikeyAdd(p1 *ecdsa.PrivateKey, p2 *ecdsa.PrivateKey) *ecdsa.PrivateKey {
	k := new(big.Int).Add(p1.D, p2.D)
	k.Mod(k, p1.Params().N)
	return NewPrikey(k)
}

func NewPrikey(d *big.Int) *ecdsa.PrivateKey {
	ret := &ecdsa.PrivateKey{}
	ret.Curve = S256()
	ret.D = new(big.Int).Set(d)
	ret.X, ret.Y = ret.ScalarBaseMult(ret.D.Bytes())
	return ret
}

func PubkeyMul(p ecdsa.PublicKey, sc *big.Int) ecdsa.PublicKey {
	ret := p
	ret.X, ret.Y = p.ScalarMult(p.X, p.Y, sc.Bytes())
	return ret
}

func PrikeyMul(p *ecdsa.PrivateKey, sc *big.Int) *ecdsa.PrivateKey {
	k := new(big.Int).Mul(p.D, sc)
	k.Mod(k, p.Params().N)
	return NewPrikey(k)
}

func fermatInverse(k, N *big.Int) *big.Int {
	two := big.NewInt(2)
	nMinus2 := new(big.Int).Sub(N, two)
	return new(big.Int).Exp(k, nMinus2, N)
}

func hashToInt(hash []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

func vrfSignX(priv *ecdsa.PrivateKey, hash []byte) (S *big.Int) {
	c := priv.Curve
	N := c.Params().N

	var r *ecdsa.PrivateKey
	R := new(big.Int)
	for h := hash; R.Sign() == 0; {
		h = Keccak256(h)
		hg := new(big.Int).SetBytes(h)
		r = PrikeyMul(priv, hg)
		R.Mod(r.X, N)
	}

	h := Keccak256(hash, priv.X.Bytes(), priv.Y.Bytes())
	x := new(big.Int).SetBytes(h)
	x.Mul(x, priv.D)
	x.Add(x, r.D)
	return x.Mod(x, N)
}

func vrfVerifyX(pub *ecdsa.PublicKey, hash []byte, S *big.Int) bool {
	c := pub.Curve
	N := c.Params().N

	var rx ecdsa.PublicKey
	R := new(big.Int)
	for h := hash; R.Sign() == 0; {
		h = Keccak256(h)
		hg := new(big.Int).SetBytes(h)
		rx = PubkeyMul(*pub, hg)
		R.Mod(rx.X, N)
	}
	h := Keccak256(hash, pub.X.Bytes(), pub.Y.Bytes())
	x := new(big.Int).SetBytes(h)
	cx := PubkeyMul(*pub, x)
	csk := PubkeyAdd(cx, rx)
	sk := NewPrikey(S)
	return csk.X.Cmp(sk.X) == 0 && csk.Y.Cmp(sk.Y) == 0
}

func vrfSign(priv *ecdsa.PrivateKey, hash []byte) (S *big.Int) {
	c := priv.Curve
	N := c.Params().N

	var invK *big.Int
	R := new(big.Int)
	for {
		hg := new(big.Int).SetBytes(Keccak256(hash))
		r := PrikeyMul(priv, hg)
		k := r.D
		R.Mod(r.X, N)
		if R.Sign() != 0 {
			invK = fermatInverse(k, N) // N != 0
			break
		}
	}

	e := hashToInt(hash, c)
	S = new(big.Int).Mul(priv.D, R)
	S.Add(S, e)
	S.Mul(S, invK)
	S.Mod(S, N) // N != 0
	if S.Sign() != 0 {
		return S
	}
	return nil
}

func vrfVerify(pub *ecdsa.PublicKey, hash []byte, S *big.Int) bool {
	c := pub.Curve
	N := c.Params().N
	RX := new(big.Int)
	for {
		hg := new(big.Int).SetBytes(Keccak256(hash))
		PubR := PubkeyMul(*pub, hg)
		RX.Mod(PubR.X, N)
		if RX.Sign() != 0 {
			break
		}
	}
	return ecdsa.Verify(pub, hash, RX, S)
}

// VRFGenerate simulate VRF
func VRF_Proof(priv *ecdsa.PrivateKey, info []byte) (proof []byte) {
	if s := vrfSignX(priv, info); s != nil {
		return s.Bytes()
	}
	return nil
}

func VRF_Verify(pub *ecdsa.PublicKey, info, proof []byte) bool {
	return vrfVerifyX(pub, info, new(big.Int).SetBytes(proof))
}

var suite = string([]byte{0x01})

func ECVRF_hash_to_curve(pub *ecdsa.PublicKey, alpha string) *ecdsa.PublicKey {
	//hash := Keccak256([]byte(alpha))
	return nil
}

func ECVRF_nonce_generation(sk *ecdsa.PrivateKey, h string) *big.Int {
	clen := (len(h) + 7) / 8
	V := make([]byte, 0, clen)
	K := make([]byte, 0, clen)
	for i := range V {
		V[i] = 0x1
		K[i] = 0
	}
	HMAC_K := func(p ...[]byte) []byte {
		p = append([][]byte{K}, p...)
		return Keccak256(p...)
	}
	K = HMAC_K(V, []byte{0x0}, sk.D.Bytes(), []byte(h))
	V = HMAC_K(V)
	K = HMAC_K(V, []byte{0x1}, sk.D.Bytes(), []byte(h))
	V = HMAC_K(V)
	return nil
}
