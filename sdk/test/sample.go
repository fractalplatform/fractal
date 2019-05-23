package main

import (
	"encoding/json"
	"math/big"

	"github.com/fractalplatform/fractal/consensus/dpos"

	"github.com/fractalplatform/fractal/accountmanager"
	"github.com/fractalplatform/fractal/common"
)

func sampleCreateAccount() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "createaccount",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
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
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleUpdateAccount() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "updateaccount",
		From:    "sampleact",
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   big.NewInt(0),
		Payload: &accountmanager.UpdataAccountAction{
			Founder: "sampleact",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleUpdateAccountAuthor() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "updateaccountauthor",
		From:    "sampleact",
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
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
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleIssueAsset() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    "sampleact",
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
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
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleUpdateAsset() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		Payload: &accountmanager.UpdateAsset{
			AssetID: 1,
			Founder: "sampleact",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleSetAssetOwner() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		Payload: &accountmanager.UpdateAssetOwner{
			AssetID: 1,
			Owner:   "sampleact",
		},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleDestroyAsset() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: 1,
		Value:   new(big.Int).Mul(big.NewInt(100000000000), big.NewInt(1e18)),
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleIncreaseAsset() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		Payload: &accountmanager.IncAsset{
			AssetId: 1,
			Amount:  new(big.Int).Mul(big.NewInt(100000000000), big.NewInt(1e18)),
		},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleTransfer() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)),
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleRegCandidate() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
		Payload: &dpos.RegisterCandidate{},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleUpdateCandidate() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
		Payload: &dpos.UpdateCandidate{},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleUnRegCandidate() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleRefundCandidate() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleVoteCandidate() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
		Payload: &dpos.VoteCandidate{},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleKickedCandidate() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
		Payload: &dpos.KickedCandidate{},
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
func sampleExitTakeOver() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value: big.NewInt(0),
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
