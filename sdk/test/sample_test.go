package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestSample(t *testing.T) {
	txs := []*TTX{
		sampleCreateAccount(),
		sampleUpdateAccount(),
		sampleUpdateAccountAuthor(),
		sampleIssueAsset(),
		sampleUpdateAsset(),
		sampleSetAssetOwner(),
		sampleDestroyAsset(),
		sampleIncreaseAsset(),
		sampleTransfer(),
		sampleRegCandidate(),
		sampleUpdateCandidate(),
		sampleUnRegCandidate(),
		sampleRefundCandidate(),
		sampleVoteCandidate(),
		sampleKickedCandidate(),
		sampleExitTakeOver(),
	}

	bts, _ := json.Marshal(txs)
	ioutil.WriteFile("sample.json", bts, 0666)

	for _, tx := range txs {
		if err := runTx(api, tx, ""); err != nil {
			panic(err)
		}
	}
}
