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

package sdk

import (
	"fmt"
	"time"

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
			r, _ := acc.api.GetTransactionReceiptByHash(hash)
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

func (acc *Account) checkCreateAccount(action *types.Action) (func() error, error) {
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) checkUpdateAccount(action *types.Action) (func() error, error) {
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) checkTranfer(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) checkIssueAsset(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) checkUpdateAsset(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) checkDestroyAsset(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) checkIncreaseAsset(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}

func (acc *Account) checkSetAssetOwner(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
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

func (acc *Account) chekRefundProdoucer(action *types.Action) (func() error, error) {
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

func (acc *Account) chekKickedCandidate(action *types.Action) (func() error, error) {
	// TODO
	function := func() error {
		return nil
	}
	return function, nil
}
