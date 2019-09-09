package rpc

import (
	"github.com/fractalplatform/fractal/common"
	jww "github.com/spf13/jwalterweatherman"
	"strconv"
)

func GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := ClientCall("ft_getBlockByHash", &block, hash.Hex(), fullTx)

	return block, err
}

func GetBlockByNumber(blockNr int64, fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := ClientCall("ft_getBlockByNumber", &block, blockNr, fullTx)

	return block, err
}

func GetCurrentBlock(fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := ClientCall("ft_getCurrentBlock", &block, fullTx)

	return block, err
}

func GetCurrentBlockHeight() (int64, error) {
	block, err := GetCurrentBlock(false)
	if err != nil {
		return -1, err
	}

	number, err := strconv.ParseInt(strconv.FormatFloat(block["number"].(float64), 'f', -1, 64), 10, 64)
	if err != nil {
		return -1, err
	}
	return number, nil
}

func GetCurrentBlockTxNum() (int64, int, error) {
	block, err := GetCurrentBlock(false)
	if err != nil {
		return -1, -1, err
	}
	txList := block["transactions"].([]interface{})

	number, err := strconv.ParseInt(strconv.FormatFloat(block["number"].(float64), 'f', -1, 64), 10, 64)
	if err != nil {
		return -1, -1, err
	}

	return number, len(txList), nil
}

func GetTxNumByBlockHeight(blockNr int64) (int, error) {
	block, err := GetBlockByNumber(blockNr, false)
	if err != nil {
		jww.INFO.Printf("GetBlockByNumber err:" + err.Error())
		return -1, err
	}
	if block == nil {
		return -1, nil
	}

	if txList, ok := block["transactions"].([]interface{}); ok {
		return len(txList), nil
	}
	return -1, nil
}
