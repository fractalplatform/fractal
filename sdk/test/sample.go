package main

import (
	"encoding/json"

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
		Value:   100,
		Payload: &accountmanager.CreateAccountAction{
			AccountName: "sdktest005",
			Founder:     "sdktest005",
			PublicKey:   common.HexToPubKey("047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd"),
			Description: "descr sdktest001",
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
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   0,
		Payload: &accountmanager.UpdataAccountAction{
			Founder: "sdktest005",
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
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   0,
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
func sampleTransfer() string {
	tx := &TTX{
		Priv:    "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032",
		Type:    "transfer",
		From:    chainCfg.SysName,
		To:      chainCfg.AccountName,
		Gas:     1000000,
		AssetID: chainCfg.SysTokenID,
		Value:   100,
		Succeed: true,
		Childs:  []*TTX{},
	}
	bts, _ := json.Marshal(tx)
	return string(bts)
}
