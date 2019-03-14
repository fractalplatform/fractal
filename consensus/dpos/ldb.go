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

package dpos

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"sort"
	"strings"

	"github.com/fractalplatform/fractal/utils/rlp"
)

// IDatabase level db
type IDatabase interface {
	Get(key string) ([]byte, error)
	Put(key string, value []byte) error
	Delete(key string) error

	Delegate(string, *big.Int) error
	Undelegate(string, *big.Int) error
	IncAsset2Acct(string, string, *big.Int) error

	GetSnapshot(key string, timestamp uint64) ([]byte, error)
}

var (
	// ProducerKeyPrefix producer name --> producerInfo
	ProducerKeyPrefix = "prod"
	// VoterKeyPrefix voter name ---> voterInfo
	VoterKeyPrefix = "vote"
	// // DelegatorKeyPrfix producer name ----> voter names
	// DelegatorKeyPrfix = "dele"
	// ProducersKeyPrefix produces ----> producer names
	ProducersKeyPrefix = "prods"
	// StateKeyPrefix height --> globalState
	StateKeyPrefix = "state"
	// Separator Split characters
	Separator = "_"

	// ProducersKey producers
	ProducersKey = "prods"
	// ProducersSizeKey producers size
	ProducersSizeKey = "prodsize"
	// LastestStateKey lastest
	LastestStateKey = "lastest"
)

// LDB dpos level db
type LDB struct {
	IDatabase
}

var _ IDB = &LDB{}

func NewLDB(db IDatabase) (*LDB, error) {
	ldb := &LDB{
		IDatabase: db,
	}
	return ldb, nil
}

func (db *LDB) GetProducer(name string) (*producerInfo, error) {
	key := strings.Join([]string{ProducerKeyPrefix, name}, Separator)
	producerInfo := &producerInfo{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(val, producerInfo); err != nil {
		return nil, err
	}
	return producerInfo, nil
}

func (db *LDB) GetVoter(name string) (*voterInfo, error) {
	key := strings.Join([]string{VoterKeyPrefix, name}, Separator)
	voterInfo := &voterInfo{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(val, voterInfo); err != nil {
		return nil, err
	}
	return voterInfo, nil
}

// func (db *LDB) GetDelegators(producer string) ([]string, error) {
// 	key := strings.Join([]string{DelegatorKeyPrfix, producer}, Separator)
// 	delegators := []string{}
// 	if val, err := db.Get(key); err != nil {
// 		return nil, err
// 	} else if val == nil {
// 		return nil, nil
// 	} else if err := rlp.DecodeBytes(val, &delegators); err != nil {
// 		return nil, err
// 	}
// 	return delegators, nil
// }

func (db *LDB) SetProducer(producer *producerInfo) error {
	key := strings.Join([]string{ProducerKeyPrefix, producer.Name}, Separator)
	if val, err := rlp.EncodeToBytes(producer); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}

	// producers
	producers := []string{}
	pkey := strings.Join([]string{ProducersKeyPrefix, ProducersKey}, Separator)
	if pval, err := db.Get(pkey); err != nil {
		return err
	} else if pval == nil {

	} else if err := rlp.DecodeBytes(pval, &producers); err != nil {
		return err
	}

	for _, name := range producers {
		if strings.Compare(name, producer.Name) == 0 {
			return nil
		}
	}

	producers = append(producers, producer.Name)
	npval, err := rlp.EncodeToBytes(producers)
	if err != nil {
		return err
	}
	if err := db.Put(pkey, npval); err != nil {
		return err
	}

	skey := strings.Join([]string{ProducersKeyPrefix, ProducersSizeKey}, Separator)
	return db.Put(skey, uint64tobytes(uint64(len(producers))))
}

func (db *LDB) SetVoter(voter *voterInfo) error {
	key := strings.Join([]string{VoterKeyPrefix, voter.Name}, Separator)
	if val, err := rlp.EncodeToBytes(voter); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}

	// 	// delegators
	// 	delegators := []string{}
	// 	dkey := strings.Join([]string{DelegatorKeyPrfix, voter.Producer}, Separator)

	// 	if dval, err := db.Get(dkey); err != nil {
	// 		return err
	// 	} else if dval == nil {

	// 	} else if err := rlp.DecodeBytes(dval, &delegators); err != nil {
	// 		return err
	// 	}
	// 	for _, name := range delegators {
	// 		if strings.Compare(name, voter.Name) == 0 {
	// 			return nil
	// 		}
	// 	}
	// 	ndval, err := rlp.EncodeToBytes(append(delegators, voter.Name))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return db.Put(dkey, ndval)
	return nil
}

func (db *LDB) DelProducer(name string) error {
	key := strings.Join([]string{ProducerKeyPrefix, name}, Separator)
	if err := db.Delete(key); err != nil {
		return err
	}

	// producers
	producers := []string{}
	pkey := strings.Join([]string{ProducersKeyPrefix, ProducersKey}, Separator)
	if pval, err := db.Get(pkey); err != nil {
		return err
	} else if pval == nil {

	} else if err := rlp.DecodeBytes(pval, &producers); err != nil {
		return err
	}
	for index, prod := range producers {
		if strings.Compare(prod, name) == 0 {
			producers = append(producers[:index], producers[index+1:]...)
			break
		}
	}

	skey := strings.Join([]string{ProducersKeyPrefix, ProducersSizeKey}, Separator)
	if err := db.Put(skey, uint64tobytes(uint64(len(producers)))); err != nil {
		return err
	}

	if len(producers) == 0 {
		return db.Delete(pkey)
	}
	npval, err := rlp.EncodeToBytes(producers)
	if err != nil {
		return err
	}
	return db.Put(pkey, npval)
}

func (db *LDB) DelVoter(name string, producer string) error {
	key := strings.Join([]string{VoterKeyPrefix, name}, Separator)
	if err := db.Delete(key); err != nil {
		return err
	}

	// delegators
	// delegators := []string{}
	// dkey := strings.Join([]string{DelegatorKeyPrfix, producer}, Separator)
	// if dval, err := db.Get(dkey); err != nil {
	// 	return err
	// } else if dval == nil {

	// } else if err := rlp.DecodeBytes(dval, &delegators); err != nil {
	// 	return err
	// }
	// for index, voter := range delegators {
	// 	if strings.Compare(voter, name) == 0 {
	// 		delegators = append(delegators[:index], delegators[index+1:]...)
	// 		break
	// 	}
	// }
	// if len(delegators) == 0 {
	// 	return db.Delete(dkey)
	// }
	// ndval, err := rlp.EncodeToBytes(delegators)
	// if err != nil {
	// 	return err
	// }
	// return db.Put(dkey, ndval)
	return nil
}

func (db *LDB) Producers() ([]*producerInfo, error) {
	// producers
	pkey := strings.Join([]string{ProducersKeyPrefix, ProducersKey}, Separator)
	producers := []string{}
	if pval, err := db.Get(pkey); err != nil {
		return nil, err
	} else if pval == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(pval, &producers); err != nil {
		return nil, err
	}

	prods := producerInfoArray{}
	for _, producer := range producers {
		prod, err := db.GetProducer(producer)
		if err != nil {
			return nil, err
		}
		prods = append(prods, prod)
	}
	sort.Sort(prods)
	return prods, nil
}

func (db *LDB) ProducersSize() (uint64, error) {
	size := uint64(0)
	skey := strings.Join([]string{ProducersKeyPrefix, ProducersSizeKey}, Separator)
	if sval, err := db.Get(skey); err != nil {
		return 0, err
	} else if sval != nil {
		size = bytestouint64(sval)
	}
	return size, nil
}

func (db *LDB) GetState(height uint64) (*globalState, error) {
	if height == LastBlockHeight {
		var err error
		height, err = db.lastestHeight()
		if err != nil {
			return nil, err
		}
	}
	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(height))}, Separator)
	gstate := &globalState{}
	if val, err := db.Get(key); err != nil {
		return nil, err
	} else if val == nil {
		return nil, nil
	} else if err := rlp.DecodeBytes(val, gstate); err != nil {
		return nil, err
	}
	return gstate, nil
}

func (db *LDB) SetState(gstate *globalState) error {
	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(gstate.Height))}, Separator)
	if val, err := rlp.EncodeToBytes(gstate); err != nil {
		return err
	} else if err := db.Put(key, val); err != nil {
		return err
	}

	// if height, err := db.lastestHeight(); err != nil {
	// 	return err
	// } else if height > gstate.Height {
	// 	panic("not reached")
	// }

	lkey := strings.Join([]string{StateKeyPrefix, LastestStateKey}, Separator)
	if err := db.Put(lkey, uint64tobytes(gstate.Height)); err != nil {
		return err
	}
	return nil
}
func (db *LDB) DelState(height uint64) error {
	if height == LastBlockHeight {
		var err error
		height, err = db.lastestHeight()
		if err != nil {
			return err
		}
	}
	key := strings.Join([]string{StateKeyPrefix, hex.EncodeToString(uint64tobytes(height))}, Separator)
	return db.Delete(key)
}

func (db *LDB) GetDelegatedByTime(name string, timestamp uint64) (*big.Int, error) {
	key := strings.Join([]string{ProducerKeyPrefix, name}, Separator)
	val, err := db.GetSnapshot(key, timestamp)
	if err != nil {
		return nil, err
	}
	if val != nil {
		producerInfo := &producerInfo{}
		if err := rlp.DecodeBytes(val, producerInfo); err != nil {
			return nil, err
		}
		return producerInfo.Quantity, nil
	}

	key = strings.Join([]string{VoterKeyPrefix, name}, Separator)
	val, err = db.GetSnapshot(key, timestamp)
	if err != nil {
		return nil, err
	}
	if val != nil {
		voterInfo := &voterInfo{}
		if err := rlp.DecodeBytes(val, voterInfo); err != nil {
			return nil, err
		}
		return voterInfo.Quantity, nil
	}
	return big.NewInt(0), nil
}

func (db *LDB) lastestHeight() (uint64, error) {
	lkey := strings.Join([]string{StateKeyPrefix, LastestStateKey}, Separator)
	if val, err := db.Get(lkey); err != nil {
		return 0, err
	} else if val == nil {
		return 0, nil
	} else {
		return bytestouint64(val), nil
	}
}

func uint64tobytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, i)
	return buf
}

func bytestouint64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}
