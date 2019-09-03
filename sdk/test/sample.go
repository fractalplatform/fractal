package main

import (
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/consensus/dpos"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
)

func sampleCreateAccount() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "createaccount",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		Payload: &accountmanager.CreateAccountAction{
			AccountName: "sampleact",
			Founder:     "sampleact",
			PublicKey:   common.HexToPubKey("047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"),
			Description: "sample account",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}
	return tx
}
func sampleUpdateAccount() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "updateaccount",
		From:    "sampleact",
		To:      chainCfg.AccountName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &accountmanager.UpdataAccountAction{
			Founder: "sampleact",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleUpdateAccountAuthor() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "updateaccountauthor",
		From:    "sampleact",
		To:      chainCfg.AccountName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &accountmanager.AccountAuthorAction{
			Threshold:             1,
			UpdateAuthorThreshold: 2,
			AuthorActions: []*accountmanager.AuthorAction{
				&accountmanager.AuthorAction{
					ActionType: accountmanager.AddAuthor,
					Author: &common.Author{
						Owner:  common.PubKey{},
						Weight: 1,
					},
				},
				&accountmanager.AuthorAction{
					ActionType: accountmanager.UpdateAuthor,
					Author: &common.Author{
						Owner:  common.PubKey{},
						Weight: 1,
					},
				},
				&accountmanager.AuthorAction{
					ActionType: accountmanager.DeleteAuthor,
					Author: &common.Author{
						Owner:  common.PubKey{},
						Weight: 1,
					},
				},
			},
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleIssueAsset() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "issueasset",
		From:    "sampleact",
		To:      chainCfg.AssetName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &accountmanager.IssueAsset{
			AssetName:   "sampleast",
			Symbol:      "sast",
			Amount:      new(big.Int).Mul(big.NewInt(100000000000), big.NewInt(1e18)),
			Decimals:    18,
			Founder:     "sampleact",
			Owner:       "sampleact",
			UpperLimit:  new(big.Int).Mul(big.NewInt(200000000000), big.NewInt(1e18)),
			Contract:    common.StrToName(""),
			Description: "sample asset",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleUpdateAsset() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "updateasset",
		From:    "sampleact",
		To:      chainCfg.AssetName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &accountmanager.UpdateAsset{
			AssetID: 1,
			Founder: "sampleact",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleSetAssetOwner() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "setassetowner",
		From:    "sampleact",
		To:      chainCfg.AssetName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &accountmanager.UpdateAssetOwner{
			AssetID: 1,
			Owner:   "sampleact",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleDestroyAsset() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "destroyasset",
		From:    "sampleact",
		To:      chainCfg.AssetName,
		Gas:     30000000,
		AssetID: 1,
		Value:   new(big.Int).Mul(big.NewInt(100000000000), big.NewInt(1e18)),
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleIncreaseAsset() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "increaseasset",
		From:    "sampleact",
		To:      chainCfg.AssetName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &accountmanager.IncAsset{
			AssetID: 1,
			Amount:  new(big.Int).Mul(big.NewInt(100000000000), big.NewInt(1e18)),
			To:      "sampleact",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleTransfer() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      "sampleact",
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleRegCandidate() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "regcandidate",
		From:    chainCfg.SysName,
		To:      chainCfg.DposName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &dpos.RegisterCandidate{
			URL: fmt.Sprintf("www.xxxxxx.com"),
		},
		Succeed: false,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleUpdateCandidate() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "updatecandidate",
		From:    chainCfg.SysName,
		To:      chainCfg.DposName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &dpos.UpdateCandidate{
			URL: fmt.Sprintf("www.xxxxxx.com"),
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleUnRegCandidate() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "unregcandidate",
		From:    "sampleact",
		To:      chainCfg.DposName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Succeed: false,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleRefundCandidate() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "refundcandidate",
		From:    chainCfg.SysName,
		To:      chainCfg.DposName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Succeed: false,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleVoteCandidate() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "votecandidate",
		From:    chainCfg.SysName,
		To:      chainCfg.DposName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &dpos.VoteCandidate{
			Candidate: chainCfg.SysName,
			Stake:     new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18)),
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleKickedCandidate() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "kickedcandidate",
		From:    chainCfg.SysName,
		To:      chainCfg.DposName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &dpos.KickedCandidate{
			Candidates: []string{
				"candidate",
			},
		},
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
func sampleExitTakeOver() *TTX {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "exittakeover",
		From:    chainCfg.SysName,
		To:      chainCfg.DposName,
		Gas:     30000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Succeed: true,
		Childs:  []*TTX{},
	}

	return tx
}
