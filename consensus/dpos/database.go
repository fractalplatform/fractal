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
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

// IDB dpos database
type IDB interface {
	SetCandidate(*CandidateInfo) error
	DelCandidate(uint64, string) error
	GetCandidate(uint64, string) (*CandidateInfo, error)
	GetCandidates(uint64) (CandidateInfoArray, error)
	CandidatesSize(uint64) (uint64, error)

	SetActivatedCandidate(uint64, *CandidateInfo) error
	GetActivatedCandidate(uint64) (*CandidateInfo, error)

	SetAvailableQuantity(uint64, string, *big.Int) error
	GetAvailableQuantity(uint64, string) (*big.Int, error)

	SetVoter(*VoterInfo) error
	GetVoter(uint64, string, string) (*VoterInfo, error)
	GetVotersByVoter(uint64, string) ([]*VoterInfo, error)
	GetVotersByCandidate(uint64, string) ([]*VoterInfo, error)

	SetState(*GlobalState) error
	GetState(uint64) (*GlobalState, error)
	SetLastestEpoch(uint64) error
	GetLastestEpoch() (uint64, error)

	SetTakeOver(uint64) error
	GetTakeOver() (uint64, error)

	Undelegate(string, *big.Int) (*types.Action, error)
	IncAsset2Acct(string, string, *big.Int, uint64) (*types.Action, error)
	GetBalanceByTime(name string, timestamp uint64) (*big.Int, error)
	GetCandidateInfoByTime(epoch uint64, name string, timestamp uint64) (*CandidateInfo, error)

	CanMine(name string, pub []byte) error
}

// CandidateType candidate status
type CandidateType uint64

const (
	// Normal reg
	Normal CandidateType = iota
	// Freeze unreg bug not del
	Freeze
	// Black in black list
	Black
	// Jail in jail list
	Jail
	// Unkown not support
	Unkown
)

// MarshalText returns the hex representation of a. Implements encoding.TextMarshaler
// is supported by most codec implementations (e.g. for yaml or toml).
func (t CandidateType) MarshalText() ([]byte, error) {
	return t.MarshalJSON()
}

// MarshalJSON returns the hex representation of a.
func (t CandidateType) MarshalJSON() ([]byte, error) {
	str := "unkown"
	switch t {
	case Normal:
		str = "normal"
	case Freeze:
		str = "freeze"
	case Black:
		str = "black"
	case Jail:
		str = "jail"

	}
	return json.Marshal(str)
}

// UnmarshalText parses a hash in syntax.
func (t *CandidateType) UnmarshalText(input []byte) error {
	return t.UnmarshalJSON(input)
}

// UnmarshalJSON parses a type in syntax.
func (t *CandidateType) UnmarshalJSON(data []byte) error {
	var val string
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	switch strings.ToLower(val) {
	case "normal":
		*t = Normal
	case "freeze":
		*t = Freeze
	case "black":
		*t = Black
	case "jail":
		*t = Jail
	default:
		*t = Unkown
	}
	return nil
}

// CandidateInfo info
type CandidateInfo struct {
	Epoch         uint64        `json:"epoch"`
	Name          string        `json:"name"`          // candidate name
	Info          string        `json:"info"`          // candidate url
	Quantity      *big.Int      `json:"quantity"`      // candidate stake quantity
	TotalQuantity *big.Int      `json:"totalQuantity"` // candidate total stake quantity
	Number        uint64        `json:"number"`        // timestamp
	Counter       uint64        `json:"shouldCounter"`
	ActualCounter uint64        `json:"actualCounter"`
	Type          CandidateType `json:"type"`
	PrevKey       string        `json:"-"`
	NextKey       string        `json:"-"`
	PubKey        common.PubKey `json:"pubkey" rlp:"-"`
}

func (candidateInfo *CandidateInfo) copy() *CandidateInfo {
	return &CandidateInfo{
		Epoch:         candidateInfo.Epoch,
		Name:          candidateInfo.Name,
		Info:          candidateInfo.Info,
		Quantity:      candidateInfo.Quantity,
		TotalQuantity: candidateInfo.TotalQuantity,
		Number:        candidateInfo.Number,
		Counter:       candidateInfo.Counter,
		ActualCounter: candidateInfo.ActualCounter,
		Type:          candidateInfo.Type,
		PubKey:        candidateInfo.PubKey,
	}
}

func (candidateInfo *CandidateInfo) invalid() bool {
	return candidateInfo.Type != Normal
}

// VoterInfo info
type VoterInfo struct {
	Epoch               uint64   `json:"epoch"`
	Name                string   `json:"name"`      // voter name
	Candidate           string   `json:"candidate"` // candidate approved by this voter
	Quantity            *big.Int `json:"quantity"`  // stake approved by this voter
	Number              uint64   `json:"number"`    // timestamp
	NextKeyForVoter     string   `json:"-"`
	NextKeyForCandidate string   `json:"-"`
}

func (voter *VoterInfo) key() string {
	return fmt.Sprintf("0x%x_%s_%s", voter.Epoch, voter.Name, voter.Candidate)
}

var (
	// InvalidIndex magic number
	InvalidIndex = uint64(math.MaxUint64)
)

// GlobalState dpos state
type GlobalState struct {
	Epoch                       uint64   `json:"epoch"`                       // epoch
	PreEpoch                    uint64   `json:"preEpoch"`                    // epoch
	ActivatedCandidateSchedule  []string `json:"activatedCandidateSchedule"`  // candidates
	ActivatedTotalQuantity      *big.Int `json:"activatedTotalQuantity"`      // the sum of activate candidate votes
	BadCandidateIndexSchedule   []uint64 `json:"badCandidateIndexSchedule"`   // activated backup candidates
	UsingCandidateIndexSchedule []uint64 `json:"usingCandidateIndexSchedule"` // activated backup candidates
	TotalQuantity               *big.Int `json:"totalQuantity"`               // the sum of all candidate votes
	TakeOver                    bool     `json:"takeOver"`                    // systemio take over dpos
	Dpos                        bool     `json:"dpos"`                        // dpos status
	Number                      uint64   `json:"number"`                      // timestamp
}

// ArrayCandidateInfoForBrowser dpos state
type ArrayCandidateInfoForBrowser struct {
	Data                        []*CandidateInfoForBrowser `json:"data"`
	BadCandidateIndexSchedule   []uint64                   `json:"bad"`
	UsingCandidateIndexSchedule []uint64                   `json:"using"`
	TakeOver                    bool                       `json:"takeOver"`
	Dpos                        bool                       `json:"dpos"`
}

// CandidateInfoForBrowser dpos state
type CandidateInfoForBrowser struct {
	Candidate        string `json:"candidate"`
	Holder           string `json:"holder"`
	Quantity         string `json:"quantity"`
	TotalQuantity    string `json:"totalQuantity"`
	Counter          uint64 `json:"shouldCounter"`
	ActualCounter    uint64 `json:"actualCounter"`
	NowCounter       uint64 `json:"nowShouldCounter"`
	NowActualCounter uint64 `json:"nowActualCounter"`
	// Info              string `json:"info"`
	// Status           uint64 `json:"status"` //0:die 1:activate 2:spare
}

type VoterInfoFractal struct {
	Candidate     string `json:"candidate"`
	Holder        string `json:"holder"`
	Quantity      string `json:"quantity"`
	TotalQuantity string `json:"totalQuantity"`
	Info          string `json:"info"`
	State         uint64 `json:"state"`
	Vote          uint64 `json:"vote"`
	CanVote       bool   `json:"canVote"`
}

// CandidateInfoArray array of candidate
type CandidateInfoArray []*CandidateInfo

// Epochs array of epcho
type Epochs struct {
	Data []*Epoch `json:"data"`
}

// Epoch timestamp & epoch number
type Epoch struct {
	Start uint64 `json:"start"`
	Epoch uint64 `json:"epoch"`
}

// VoteEpochs array of epcho
type VoteEpochs struct {
	Data []*VoteEpoch `json:"data"`
}

// VoteEpoch timestamp & epoch number & dpos status
type VoteEpoch struct {
	Start uint64 `json:"start"`
	Epoch uint64 `json:"epoch"`
	Dpos  uint64 `json:"dpos"`
}

func (prods CandidateInfoArray) Len() int {
	return len(prods)
}
func (prods CandidateInfoArray) Less(i, j int) bool {
	return more(prods[i], prods[j])
}
func (prods CandidateInfoArray) Swap(i, j int) {
	prods[i], prods[j] = prods[j], prods[i]
}

func more(frist *CandidateInfo, second *CandidateInfo) bool {
	val := frist.TotalQuantity.Cmp(second.TotalQuantity)
	if val == 0 {
		if frist.Number == second.Number {
			return strings.Compare(frist.Name, second.Name) > 0
		}
		return frist.Number < second.Number
	}
	return val > 0
}
