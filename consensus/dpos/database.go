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
	"fmt"
	"math/big"
	"strings"
)

// IDB dpos database
type IDB interface {
	SetCandidate(*CandidateInfo) error
	DelCandidate(string) error
	GetCandidate(string) (*CandidateInfo, error)
	GetCandidates() ([]string, error)
	CandidatesSize() (uint64, error)

	SetAvailableQuantity(uint64, string, *big.Int) error
	GetAvailableQuantity(uint64, string) (*big.Int, error)

	SetVoter(*VoterInfo) error
	DelVoter(*VoterInfo) error
	DelVoters(uint64, string) error
	GetVoter(uint64, string, string) (*VoterInfo, error)
	GetVoters(uint64, string) ([]string, error)
	GetVoterCandidates(uint64, string) ([]string, error)

	SetState(*GlobalState) error
	GetState(uint64) (*GlobalState, error)

	Undelegate(string, *big.Int) error
	IncAsset2Acct(string, string, *big.Int) error

	GetDelegatedByTime(string, uint64) (*big.Int, *big.Int, uint64, error)
}

// CandidateInfo info
type CandidateInfo struct {
	Name          string   `json:"name"`          // candidate name
	URL           string   `json:"url"`           // candidate url
	Quantity      *big.Int `json:"quantity"`      // candidate stake quantity
	TotalQuantity *big.Int `json:"totalQuantity"` // candidate total stake quantity
	Height        uint64   `json:"height"`        // timestamp
	Counter       uint64   `json:"counter"`
	InBlackList   bool     `json:"inBlackList"`
	InJail        bool     `json:"inJail"`
}

// VoterInfo info
type VoterInfo struct {
	Epcho     uint64   `json:"epcho"`
	Name      string   `json:"name"`      // voter name
	Candidate string   `json:"candidate"` // candidate approved by this voter
	Quantity  *big.Int `json:"quantity"`  // stake approved by this voter
	Height    uint64   `json:"height"`    // timestamp
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
	Epcho                            uint64   `json:"epcho"`                            // epcho
	ActivatedCandidateScheduleUpdate uint64   `json:"activatedCandidateScheduleUpdate"` // update time
	ActivatedCandidateSchedule       []string `json:"activatedCandidateSchedule"`       // candidates
	ActivatedTotalQuantity           *big.Int `json:"activatedTotalQuantity"`           // the sum of activate candidate votes
	TotalQuantity                    *big.Int `json:"totalQuantity"`                    // the sum of all candidate votes
	TakeOver                         bool     `json:"takeOver"`                         // systemio take over dpos
	Dpos                             bool     `json:"dpos"`                             // dpos status
}

type candidateInfoArray []*CandidateInfo

func (prods candidateInfoArray) Len() int {
	return len(prods)
}
func (prods candidateInfoArray) Less(i, j int) bool {
	val := prods[i].TotalQuantity.Cmp(prods[j].TotalQuantity)
	if val == 0 {
		if prods[i].Height == prods[j].Height {
			return strings.Compare(prods[i].Name, prods[j].Name) > 0
		}
		return prods[i].Height < prods[j].Height
	}
	return val > 0
}
func (prods candidateInfoArray) Swap(i, j int) {
	prods[i], prods[j] = prods[j], prods[i]
}
