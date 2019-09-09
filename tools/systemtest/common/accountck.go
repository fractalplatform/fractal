package common

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/fractalplatform/fractal/utils/rlp"

	"github.com/fractalplatform/fractal/asset"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/types"
)

func (acc *Account) utilReceipt(hash common.Hash, timeout int64) error {
	cnt := int64(10)
	ticker := time.NewTicker(time.Duration(timeout / cnt))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r, _ := acc.api.TransactionReceiptByHash(hash)
			if r != nil && r.BlockNumber > 0 {
				if len(r.ActionResults[0].Error) > 0 {
					return fmt.Errorf(r.ActionResults[0].Error)
				}
				return nil
			}
			cnt--
			if cnt == 0 {
				return fmt.Errorf("not found %v receipt", hash.String())
			}
		}
	}
}

func (acc *Account) checkTranfer(action *types.Action) (func() error, error) {
	to, err := acc.api.BalanceByAssetID(action.Recipient().String(), action.AssetID())
	if err != nil {
		return nil, err
	}

	function := func() error {
		tto, err := acc.api.BalanceByAssetID(action.Recipient().String(), action.AssetID())
		if err != nil {
			return err
		}
		if b := new(big.Int).Add(to, action.Value()); tto.Cmp(b) != 0 {
			return fmt.Errorf("after: tranfer %v err -- have: %v, except:%v", action.Recipient().String(), tto, b)
		}
		return nil
	}
	return function, nil
}

func (acc *Account) checkCreateAccount(action *types.Action) (func() error, error) {
	if action.Type() != types.CreateAccount {
		panic("mismath type")
	}
	existed, err := acc.api.AccountIsExist(action.Recipient().String())
	if err != nil {
		return nil, err
	}
	if existed {
		return nil, fmt.Errorf("before: create account %v err -- exist", action.Recipient().String())
	}

	function := func() error {
		texisted, err := acc.api.AccountIsExist(action.Recipient().String())
		if err != nil {
			return err
		}

		if !texisted {
			return fmt.Errorf("after: create account %v err -- not exist", action.Recipient().String())
		}
		return nil
	}
	return function, nil
}

func (acc *Account) checkUpdateAccount(action *types.Action) (func() error, error) {
	if action.Type() != types.UpdateAccount {
		panic("mismath type")
	}
	acct, err := acc.api.AccountInfo(action.Sender().String())
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, fmt.Errorf("before: update account %v err -- not exist", action.Sender().String())
	}

	function := func() error {
		tacct, err := acc.api.AccountInfo(action.Sender().String())
		if err != nil {
			return err
		}

		if tacct == nil {
			return fmt.Errorf("after: update account %v err -- not exist", action.Sender().String())
		}

		if strings.Compare(acct.Founder.String(), tacct.Founder.String()) != 0 {
			return fmt.Errorf("after: update account %v err -- have: %v, except:%v", action.Sender().String(), acct.Founder.String(), tacct.Founder.String())
		}
		return nil
	}
	return function, nil
}

func (acc *Account) checkDeleteAccount(action *types.Action) (func() error, error) {
	if action.Type() != types.DeleteAccount {
		panic("mismath type")
	}
	existed, err := acc.api.AccountIsExist(action.Sender().String())
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, fmt.Errorf("before: delete account %v err -- not exist", action.Sender().String())
	}

	function := func() error {
		texisted, err := acc.api.AccountIsExist(action.Sender().String())
		if err != nil {
			return err
		}

		if texisted {
			return fmt.Errorf("after: delete account %v err -- exist", action.Sender().String())
		}
		return nil
	}
	return function, nil
}

func (acc *Account) checkIssueAsset(action *types.Action) (func() error, error) {
	if action.Type() != types.IssueAsset {
		panic("mismath type")
	}
	asset := &asset.AssetObject{}
	rlp.DecodeBytes(action.Data(), asset)
	a, err := acc.api.AssetInfoByName(asset.AssetName)
	if err != nil {
		return nil, err
	}
	if a != nil {
		return nil, fmt.Errorf("before: issue asset %v err -- exist", asset.AssetName)
	}

	function := func() error {
		ta, err := acc.api.AssetInfoByName(asset.AssetName)
		if err != nil {
			return err
		}

		if ta == nil {
			return fmt.Errorf("after: issue asset %v err -- not exist", asset.AssetName)
		}
		return nil
	}
	return function, nil
}

func (acc *Account) checkIncreaseAsset(action *types.Action) (func() error, error) {
	if action.Type() != types.IncreaseAsset {
		panic("mismath type")
	}
	asset := &asset.AssetObject{}
	rlp.DecodeBytes(action.Data(), asset)
	a, err := acc.api.AssetInfoByName(asset.AssetName)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, fmt.Errorf("before: increase asset %v err -- not exist", asset.AssetName)
	}
	owerb, err := acc.api.BalanceByAssetID(a.Owner.String(), a.AssetID)
	if err != nil {
		return nil, err
	}

	function := func() error {
		ta, err := acc.api.AssetInfoByName(asset.AssetName)
		if err != nil {
			return err
		}

		if ta == nil {
			return fmt.Errorf("after: increase asset %v err -- not exist", asset.AssetName)
		}
		towerb, err := acc.api.BalanceByAssetID(a.Owner.String(), a.AssetID)
		if err != nil {
			return err
		}

		if b := new(big.Int).Add(owerb, asset.Amount); towerb.Cmp(b) != 0 {
			return fmt.Errorf("after: increase asset %v err -- have %v, except: %v ", asset.AssetName, towerb, b)
		}

		if strings.Compare(ta.Owner.String(), a.Owner.String()) != 0 {
			return fmt.Errorf("after: increase asset %v err -- have %v, except: %v ", asset.AssetName, a.Owner.String(), ta.Owner.String())
		}
		return nil
	}
	return function, nil
}

func (acc *Account) checkSetAssetOwner(action *types.Action) (func() error, error) {
	if action.Type() != types.SetAssetOwner {
		panic("mismath type")
	}
	asset := &asset.AssetObject{}
	rlp.DecodeBytes(action.Data(), asset)
	a, err := acc.api.AssetInfoByName(asset.AssetName)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, fmt.Errorf("before: increase asset %v err -- not exist", asset.AssetName)
	}

	function := func() error {
		ta, err := acc.api.AssetInfoByName(asset.AssetName)
		if err != nil {
			return err
		}

		if ta == nil {
			return fmt.Errorf("after: increase asset %v err -- not exist", asset.AssetName)
		}

		if strings.Compare(ta.Owner.String(), asset.Owner.String()) != 0 {
			return fmt.Errorf("after: increase asset %v err -- have %v, except: %v ", asset.AssetName, asset.Owner.String(), ta.Owner.String())
		}
		return nil
	}
	return function, nil
}

func (acc *Account) chekRegProdoucer(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) chekUpdateProdoucer(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) chekUnregProdoucer(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) chekVoteProdoucer(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) chekChangeProdoucer(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) chekUnvoteProdoucer(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) chekRemoveVoter(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}
