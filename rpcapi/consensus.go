package rpcapi

import (
	"github.com/fractalplatform/fractal/plugin"
)

type ConsensusAPI struct {
	Backend
}

func NewConsensusAPI(b Backend) *ConsensusAPI {
	return &ConsensusAPI{b}
}

func (api *ConsensusAPI) GetAllCandidates() ([]string, error) {
	pm, err := api.GetPM()
	if err != nil {
		return nil, err
	}
	return pm.GetAllCandidates(), nil
}

func (api *ConsensusAPI) GetCandidateInfo(account string) (*plugin.CandidateInfo, error) {
	pm, err := api.GetPM()
	if err != nil {
		return nil, err
	}
	return pm.GetCandidateInfo(account), nil
}
