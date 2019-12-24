// Copyright 2019 The Fractal Team Authors
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

package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/fractalplatform/fractal/params"
	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/types/envelope"
	"github.com/fractalplatform/fractal/utils/rlp"
)

var (
	ErrWrongTransaction = errors.New("transaction is invalid")
	ErrWrongContract    = errors.New("contract is invalid")
)

// Manager manage all plugins.
type Manager struct {
	stateDB         *state.StateDB
	contracts       map[string]IContract
	contractsByType map[envelope.PayloadType]IContract
	IAccount
	IAsset
	IConsensus
	IFee
	IItem
	ISigner
}

func init() {
	PluginSolAPIRegister(&Consensus{})
	PluginSolAPIRegister(&AccountManager{})
	PluginSolAPIRegister(&AssetManager{})
	PluginSolAPIRegister(&ItemManager{})
}

type ContextSol struct {
	pm IPM
	tx *envelope.PluginTx
}

// NewPM create new plugin manager.
func NewPM(stateDB *state.StateDB) IPM {
	acm, _ := NewACM(stateDB)
	asm, _ := NewASM(stateDB)
	consensus := NewConsensus(stateDB)
	chainID := big.NewInt(1)
	signer, _ := NewSigner(chainID)
	fee, _ := NewFeeManager()
	item, _ := NewItemManage(stateDB)
	pm := &Manager{
		contracts:       make(map[string]IContract),
		contractsByType: make(map[envelope.PayloadType]IContract),
		IAccount:        acm,
		IAsset:          asm,
		IConsensus:      consensus,
		ISigner:         signer,
		IFee:            fee,
		IItem:           item,
		stateDB:         stateDB,
	}
	pm.contracts[acm.AccountName()] = acm
	pm.contracts[asm.AccountName()] = asm
	pm.contracts[consensus.AccountName()] = consensus
	pm.contractsByType[Transfer] = acm
	pm.contracts[item.AccountName()] = item
	return pm
}

func (pm *Manager) BasicCheck(tx *types.Transaction) error {
	ptx, ok := tx.Envelope.(*envelope.PluginTx)
	if !ok {
		return ErrWrongTransaction
	}

	switch ptx.PayloadType() {
	case CreateAccount:
		param := &CreateAccountAction{}
		if err := rlp.DecodeBytes(ptx.GetPayload(), param); err != nil {
			return err
		}
		// if action.Recipient() != chainCfg.AccountName {
		// 	return fmt.Errorf("Receipt should is %v", chainCfg.AccountName)
		// }
		if ptx.Recipient() != "fractalaccount" {
			return fmt.Errorf("Receipt should is fractalaccount")
		}
		if err := pm.checkCreateAccount(param.Name, param.Pubkey, param.Desc); err != nil {
			return err
		}
	case IssueAsset:
		param := &IssueAssetAction{}
		if err := rlp.DecodeBytes(ptx.GetPayload(), param); err != nil {
			return err
		}
		// if action.Recipient() != chainCfg.AssetName {
		// 	return fmt.Errorf("Receipt should is %v", chainCfg.AssetName)
		// }
		if ptx.Recipient() != "fractalasset" {
			return fmt.Errorf("Receipt should is fractalasset")
		}
		if err := pm.checkIssueAsset(ptx.Sender(), param.AssetName, param.Symbol, param.Amount, param.Decimals, param.Founder, param.Owner, param.UpperLimit, param.Description, pm.IAccount); err != nil {
			return err
		}
	case IncreaseAsset:
		param := &IncreaseAssetAction{}
		if err := rlp.DecodeBytes(ptx.GetPayload(), param); err != nil {
			return err
		}
		// if action.Recipient() != chainCfg.AssetName {
		// 	return fmt.Errorf("Receipt should is %v", chainCfg.AssetName)
		// }
		if ptx.Recipient() != "fractalasset" {
			return fmt.Errorf("Receipt should is fractalasset")
		}
		if err := pm.checkIncreaseAsset(ptx.Sender(), param.To, param.AssetID, param.Amount, pm.IAccount); err != nil {
			return err
		}
	default:
		if ptx.PayloadType() != Transfer && (ptx.PayloadType() < RegisterMiner || ptx.PayloadType() >= ConsensusEnd) {
			return ErrWrongTransaction
		}
	}
	return nil
}

func (pm *Manager) selectContract(tx *envelope.PluginTx) IContract {
	contractName := tx.Recipient()
	if contract, exist := pm.contracts[contractName]; exist {
		return contract
	}
	return pm.contractsByType[tx.PayloadType()]
}

func (pm *Manager) ExecTx(tx *types.Transaction, fromSol bool) ([]byte, error) {
	ptx, ok := tx.Envelope.(*envelope.PluginTx)
	if !ok {
		return nil, ErrWrongTransaction
	}

	if contract := pm.selectContract(ptx); contract != nil {
		snapshot := pm.stateDB.Snapshot()
		var ret []byte
		var err error
		if fromSol {
			ret, err = PluginSolAPICall(contract, &ContextSol{pm, ptx}, ptx.Payload)
		} else {
			ret, err = contract.CallTx(ptx, pm)
		}
		if err != nil {
			pm.stateDB.RevertToSnapshot(snapshot)
		}
		return ret, err
	}
	return nil, ErrWrongContract
}

func (pm *Manager) IsPlugin(name string) bool {
	_, exist := pm.contracts[name]
	return exist
}

func (pm *Manager) InitChain(pluginDoc json.RawMessage, chainConfig *params.ChainConfig) ([]*types.Transaction, error) {
	if len(pluginDoc) == 0 {
		pluginDoc = DefaultPluginDoc()
	}

	pd, err := PluginDocJsonUnMarshal(pluginDoc)
	if err != nil {
		return nil, fmt.Errorf("Init chain json unmarshal err: %v", err)
	}

	actTxs, err := pd.CreateAccount(chainConfig.ChainName, chainConfig.AccountName)
	if err != nil {
		return nil, err
	}

	astTxs, err := pd.CreateAsset(chainConfig.ChainName, chainConfig.AssetName)
	if err != nil {
		return nil, err
	}

	minerTxs, err := pd.RegisterMiner(chainConfig.SysName, chainConfig.DposName)
	if err != nil {
		return nil, err
	}

	actTxs = append(actTxs, astTxs...)
	actTxs = append(actTxs, minerTxs...)

	for index, action := range actTxs {
		_, err := pm.ExecTx(action, false)
		if err != nil {
			return nil, fmt.Errorf("genesis index: %v, err: %v", index, err)
		}
	}

	return actTxs, nil
}
