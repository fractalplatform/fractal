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
	"math"
	"math/big"
	"strings"
)

var (
	// LastBlockHeight latest
	LastBlockHeight = uint64(math.MaxUint64)
)

// IDB dpos database
type IDB interface {
	SetCandidate(*candidateInfo) error
	DelCandidate(string) error
	GetCandidate(string) (*candidateInfo, error)
	Candidates() ([]*candidateInfo, error)
	CandidatesSize() (uint64, error)

	SetVoter(*voterInfo) error
	DelVoter(string, string) error
	GetVoter(string) (*voterInfo, error)
	// GetDelegators(string) ([]string, error)

	SetState(*globalState) error
	DelState(uint64) error
	GetState(uint64) (*globalState, error)

	Delegate(string, *big.Int) error
	Undelegate(string, *big.Int) error
	IncAsset2Acct(string, string, *big.Int) error

	GetDelegatedByTime(string, uint64) (*big.Int, *big.Int, uint64, error)
}

type candidateInfo struct {
	Name          string   `json:"name"`          // candidate name
	URL           string   `json:"url"`           // candidate url
	Quantity      *big.Int `json:"quantity"`      // candidate stake quantity
	TotalQuantity *big.Int `json:"totalQuantity"` // candidate total stake quantity
	Height        uint64   `json:"height"`        // timestamp
	Counter       uint64   `json:"counter"`
	InBlackList   bool     `json:"inBlackList"`
}

type voterInfo struct {
	Name      string   `json:"name"`      // voter name
	Candidate string   `json:"candidate"` // candidate approved by this voter
	Quantity  *big.Int `json:"quantity"`  // stake approved by this voter
	Height    uint64   `json:"height"`    // timestamp
}

type globalState struct {
	Height                           uint64   `json:"height"`                           // block height
	ActivatedCandidateScheduleUpdate uint64   `json:"activatedCandidateScheduleUpdate"` // update time
	ActivatedCandidateSchedule       []string `json:"activatedCandidateSchedule"`       // candidates
	ActivatedTotalQuantity           *big.Int `json:"activatedTotalQuantity"`           // the sum of activate candidate votes
	TotalQuantity                    *big.Int `json:"totalQuantity"`                    // the sum of all candidate votes
	TakeOver                         bool     `json:"takeOver"`                         // systemio take over dpos
}

type candidateInfoArray []*candidateInfo

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
