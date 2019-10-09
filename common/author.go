package common

import (
	"encoding/json"
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
		Owner  `json:"owner"`
		Weight uint64 `json:"weight"`
	}
	Owner interface {
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

type AuthorJSON struct {
	authorType AuthorType
	OwnerStr   string `json:"owner"`
	Weight     uint64 `json:"weight"`
}

func NewAuthor(owner Owner, weight uint64) *Author {
	return &Author{Owner: owner, Weight: weight}
}

func (a *Author) GetWeight() uint64 {
	return a.Weight
}

func (a *Author) EncodeRLP(w io.Writer) error {
	storageAuthor, err := a.encode()
	if err != nil {
		return err
	}
	return rlp.Encode(w, storageAuthor)
}

func (a *Author) encode() (*StorageAuthor, error) {
	switch aTy := a.Owner.(type) {
	case Name:
		value, err := rlp.EncodeToBytes(&aTy)
		if err != nil {
			return nil, err
		}
		return &StorageAuthor{
			Type:    AccountNameType,
			DataRaw: value,
			Weight:  a.Weight,
		}, nil
	case PubKey:
		value, err := rlp.EncodeToBytes(&aTy)
		if err != nil {
			return nil, err
		}
		return &StorageAuthor{
			Type:    PubKeyType,
			DataRaw: value,
			Weight:  a.Weight,
		}, nil
	case Address:
		value, err := rlp.EncodeToBytes(&aTy)
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
		a.Owner = name
		a.Weight = sa.Weight
		return nil
	case PubKeyType:
		var pubKey PubKey
		if err := rlp.DecodeBytes(sa.DataRaw, &pubKey); err != nil {
			return err
		}
		a.Owner = pubKey
		a.Weight = sa.Weight
		return nil
	case AddressType:
		var address Address
		if err := rlp.DecodeBytes(sa.DataRaw, &address); err != nil {
			return err
		}
		a.Owner = address
		a.Weight = sa.Weight
		return nil
	}
	return errors.New("Author decode failed")
}

func (a *Author) MarshalJSON() ([]byte, error) {
	return json.Marshal(&AuthorJSON{OwnerStr: a.Owner.String(), Weight: a.Weight})
}

func (a *Author) UnmarshalJSON(data []byte) error {
	aj := &AuthorJSON{}
	if err := json.Unmarshal(data, aj); err != nil {
		return err
	}
	a.Owner = Name(aj.OwnerStr)
	a.Weight = aj.Weight
	return nil
}
