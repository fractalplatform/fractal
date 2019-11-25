package plugin

import (
	"math/big"
	"testing"
)

func TestConsensus(t *testing.T) {
	info1 := &CandidateInfo{
		Balance: big.NewInt(100),
		Weight:  90,
	}
	info2 := &CandidateInfo{
		Balance: big.NewInt(100),
		Weight:  90,
	}
	info1.update(info2)
	if info1.Weight != 90 || info1.Balance.Cmp(big.NewInt(200)) != 0 {
		t.Fatal("wrong")
	}
}
