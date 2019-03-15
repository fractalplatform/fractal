package common

import (
	"errors"
	"io"

	"github.com/fractalplatform/fractal/utils/rlp"
)

type AuthorType uint8

const (
	AccountNameType AuthorType = iota
	PubKeyType
	AddressType
)

type (
	Author struct {
		Related
		Weight uint64 `json:"weight"`
	}
	Related interface {
		String() string
	}
)

type AccountAuthor struct {
	Account Name
}
type StorageAuthor struct {
	Type    AuthorType
	DataRaw rlp.RawValue
	Weight  uint64
}

func NewAuthor(related Related, weight uint64) *Author {
	return &Author{Related: related, Weight: weight}
}

func (a *Author) EncodeRLP(w io.Writer) error {
	storageAuthor, err := a.encode()
	if err != nil {
		return err
	}
	return rlp.Encode(w, storageAuthor)
}

func (a *Author) encode() (*StorageAuthor, error) {
	switch aTy := a.Related.(type) {
	case Name:
		value, err := rlp.EncodeToBytes(aTy)
		if err != nil {
			return nil, err
		}
		return &StorageAuthor{
			Type:    AccountNameType,
			DataRaw: value,
			Weight:  a.Weight,
		}, nil
	case PubKey:
		value, err := rlp.EncodeToBytes(aTy)
		if err != nil {
			return nil, err
		}
		return &StorageAuthor{
			Type:    PubKeyType,
			DataRaw: value,
			Weight:  a.Weight,
		}, nil
	case Address:
		value, err := rlp.EncodeToBytes(aTy)
		if err != nil {
			return nil, err
		}
		return &StorageAuthor{
			Type:    AddressType,
			DataRaw: value,
			Weight:  a.Weight,
		}, nil
	}
	return nil, errors.New("Author encode failed")
}

func (a *Author) DecodeRLP(s *rlp.Stream) error {
	storageAuthor := new(StorageAuthor)
	err := s.Decode(storageAuthor)
	if err != nil {
		return err
	}
	return a.decode(storageAuthor)
}

func (a *Author) decode(sa *StorageAuthor) error {
	switch sa.Type {
	case AccountNameType:
		var name Name
		if err := rlp.DecodeBytes(sa.DataRaw, &name); err != nil {
			return err
		}
		a.Related = name
		a.Weight = sa.Weight
		return nil
	case PubKeyType:
		var pubKey PubKey
		if err := rlp.DecodeBytes(sa.DataRaw, &pubKey); err != nil {
			return err
		}
		a.Related = pubKey
		a.Weight = sa.Weight
		return nil
	case AddressType:
		var address Address
		if err := rlp.DecodeBytes(sa.DataRaw, &address); err != nil {
			return err
		}
		a.Related = address
		a.Weight = sa.Weight
		return nil
	}
	return errors.New("Author decode failed")
}
