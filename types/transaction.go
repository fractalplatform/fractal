package types

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/log"
	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	// ErrOversizedData is returned if the input data of a transaction is greater
	// than some meaningful limit a user might use. This is not a consensus error
	// making the transaction invalid, rather a DOS protection.
	ErrOversizedData = errors.New("oversized data")
)

type Transaction struct {
	envelope.Envelope

	// caches
	pubkeys atomic.Value
	hash    atomic.Value
	size    atomic.Value
}

// NewTransaction initialize a transaction.
func NewTransaction(e envelope.Envelope) *Transaction {
	return &Transaction{
		Envelope: e,
	}
}

// Hash hashes the RLP encoding of tx.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := tx.Envelope.Hash()
	tx.hash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction,
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	tx.EncodeRLP(&c)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

type wrapper struct {
	Type envelope.Type
	Raw  rlp.RawValue
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	raw, err := rlp.EncodeToBytes(tx.Envelope)
	if err != nil {
		return err
	}
	return rlp.Encode(w, wrapper{Type: tx.Envelope.Type(), Raw: raw})
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	w := new(wrapper)

	_, size, _ := s.Kind()

	err := s.Decode(w)
	if err != nil {
		return err
	}

	tx.Envelope, err = envelope.New(w.Type)
	if err != nil {
		return err
	}

	err = rlp.DecodeBytes(w.Raw, tx.Envelope)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}
	return err
}

func (tx *Transaction) Type() envelope.Type {
	if tx == nil {
		return envelope.Unknown
	}
	return tx.Envelope.Type()
}

// Check the validity of all fields
func (tx *Transaction) Check(conf *params.ChainConfig) error {
	// Heuristic limit, reject transactions over 32KB to prfeed DOS attacks
	if tx.Size() > common.StorageSize(params.MaxTxSize) {
		return ErrOversizedData
	}
	if conf.SysTokenID != tx.GetGasAssetID() {
		return fmt.Errorf("only support system asset %d as tx fee", conf.SysTokenID)
	}
	return nil
}

// RecoverMultiKey recover and store cache.
func RecoverMultiKey(recover func(signature []byte, signHash func(chainID *big.Int) common.Hash) ([]byte, error), tx *Transaction) {
	if sc := tx.pubkeys.Load(); sc != nil {
		return
	}

	pubKey, err := recover(tx.GetSign(), tx.SignHash)
	if err != nil {
		// There should be no problem here.
		log.Error("recover failed", "err", err, "hash", tx.Hash())
	}

	tx.pubkeys.Store(pubKey)
}
