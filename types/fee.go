package types

//type for fee
const (
	AssetFeeType    = uint64(0)
	ContractFeeType = uint64(1)
	CoinbaseFeeType = uint64(2)
)

type DistributeGas struct {
	Value  int64
	TypeID uint64
}

type DistributeKey struct {
	ObjectName string
	ObjectType uint64
}

type DistributeKeys []DistributeKey

func (keys DistributeKeys) Len() int {
	return len(keys)
}

func (keys DistributeKeys) Less(i, j int) bool {
	if keys[i].ObjectName == keys[j].ObjectName {
		return keys[i].ObjectType < keys[j].ObjectType
	}
	return keys[i].ObjectName < keys[j].ObjectName
}

func (keys DistributeKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}
