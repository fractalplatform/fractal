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

	"github.com/fractalplatform/fractal/types"
)

// LastEpcho latest
var LastEpcho = uint64(math.MaxUint64)

// IDB dpos database
type IDB interface {
	SetCandidate(*CandidateInfo) error
	DelCandidate(string) error
	GetCandidate(string) (*CandidateInfo, error)
	GetCandidates() ([]*CandidateInfo, error)
	CandidatesSize() (uint64, error)

	SetAvailableQuantity(uint64, string, *big.Int) error
	GetAvailableQuantity(uint64, string) (*big.Int, error)

	SetVoter(*VoterInfo) error
	GetVoter(uint64, string, string) (*VoterInfo, error)
	GetVotersByVoter(uint64, string) ([]*VoterInfo, error)
	GetVotersByCandidate(uint64, string) ([]*VoterInfo, error)

	SetState(*GlobalState) error
	GetState(uint64) (*GlobalState, error)
	GetLastestEpcho() (uint64, error)

	Undelegate(string, *big.Int) (*types.Action, error)
	IncAsset2Acct(string, string, *big.Int) (*types.Action, error)
	GetBalanceByTime(name string, timestamp uint64) (*big.Int, error)

	GetDelegatedByTime(string, uint64) (*big.Int, *big.Int, uint64, error)
}

type CandidateType uint64

const (
	Normal CandidateType = iota
	Freeze
	Black
	Jail
	Unkown
)

// MarshalText returns the hex representation of a.
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
	Name          string        `json:"name"`          // candidate name
	URL           string        `json:"url"`           // candidate url
	Quantity      *big.Int      `json:"quantity"`      // candidate stake quantity
	TotalQuantity *big.Int      `json:"totalQuantity"` // candidate total stake quantity
	Height        uint64        `json:"height"`        // timestamp
	Counter       uint64        `json:"counter"`
	Type          CandidateType `json:"type"`
	PrevKey       string        `json:"-"`
	NextKey       string        `json:"-"`
}

func (candidateInfo *CandidateInfo) invalid() bool {
	return candidateInfo.Type != Normal
}

// VoterInfo info
type VoterInfo struct {
	Epcho               uint64   `json:"epcho"`
	Name                string   `json:"name"`      // voter name
	Candidate           string   `json:"candidate"` // candidate approved by this voter
	Quantity            *big.Int `json:"quantity"`  // stake approved by this voter
	Height              uint64   `json:"height"`    // timestamp
	NextKeyForVoter     string   `json:"-"`
	NextKeyForCandidate string   `json:"-"`
}

func (voter *VoterInfo) key() string {
	return fmt.Sprintf("0x%x_%s_%s", voter.Epcho, voter.Name, voter.Candidate)
}

func (voter *VoterInfo) ckey() string {
	return fmt.Sprintf("0x%x_%s", voter.Epcho, voter.Candidate)
}

func (voter *VoterInfo) vkey() string {
	return fmt.Sprintf("0x%x_%s", voter.Epcho, voter.Name)
}

// GlobalState dpos state
type GlobalState struct {
	Epcho                      uint64   `json:"epcho"`                      // epcho
	PreEpcho                   uint64   `json:"preEpcho"`                   // epcho
	ActivatedCandidateSchedule []string `json:"activatedCandidateSchedule"` // candidates
	ActivatedTotalQuantity     *big.Int `json:"activatedTotalQuantity"`     // the sum of activate candidate votes
	TotalQuantity              *big.Int `json:"totalQuantity"`              // the sum of all candidate votes
	TakeOver                   bool     `json:"takeOver"`                   // systemio take over dpos
	Dpos                       bool     `json:"dpos"`                       // dpos status
	Height                     uint64   `json:"height"`                     // timestamp
}

// CandidateInfoArray array of candidate
type CandidateInfoArray []*CandidateInfo

func (prods CandidateInfoArray) Len() int {
	return len(prods)
}
func (prods CandidateInfoArray) Less(i, j int) bool {
	val := prods[i].TotalQuantity.Cmp(prods[j].TotalQuantity)
	if val == 0 {
		if prods[i].Height == prods[j].Height {
			return strings.Compare(prods[i].Name, prods[j].Name) > 0
		}
		return prods[i].Height < prods[j].Height
	}
	return val > 0
}
func (prods CandidateInfoArray) Swap(i, j int) {
	prods[i], prods[j] = prods[j], prods[i]
}
