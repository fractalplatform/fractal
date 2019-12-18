package plugin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPluginDocJsonUnMarshal(t *testing.T) {
	raw := json.RawMessage([]byte(`{
    "Accounts":[
        {
            "Name":"fractalfounder",
            "Pubkey":"047db227d7094ce215c3a0f57e1bcc732551fe351f94249471934567e0f5dc1bf795962b8cccb87a2eb56b29fbe37d614e2f4c3c45b789ae4f1f51f4cb21972ffd",
            "Desc":"system account"
        },
        {
            "Name":"fractalaccount",
            "Pubkey":"0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "Desc":"account manager account"
        },
        {
            "Name":"fractalasset",
            "Pubkey":"0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "Desc":"asset manager account"
        },
        {
            "Name":"fractaldpos",
            "Pubkey":"0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "Desc":"consensus account"
        },
        {
            "Name":"fractalfee",
            "Pubkey":"0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
            "Desc":"fee manager account"
        }
    ],
    "Assets":[
        {
            "AssetName":"ftoken",
            "Symbol":"ft",
            "Amount":10000000000000000000000000000,
            "Owner":"fractalfounder",
            "Founder":"fractalfounder",
            "Decimals":18,
            "UpperLimit":10000000000000000000000000000,
            "Contract":"",
            "Description":""
        }
    ]
}`))

	_, err := PluginDocJsonUnMarshal(raw)

	assert.NoError(t, err)
}
