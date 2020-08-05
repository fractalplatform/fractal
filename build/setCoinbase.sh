# 脚本用法：
# ./setCoinbase 1 : 表示给矿工1设置coinbase
# ./setCoinbase 1 28 : 表示给矿工1~矿工28设置coinbase
# 为什么要制作此脚本：因为节点刚启动时，区块同步的少，账号也尚未创建，设置coinbase会失败，所以需要在区块同步到账号已创建的时候，才开始设置coinbase
if [[ $# -eq 0 ]]; then
        echo "command=>$0, no parameters"
        exit 1
fi

function setCoinbase ()
{
        minerName="minernodetest$1"
        ./oex miner -i ./data/$minerName/oex.ipc setcoinbase "$minerName" keys/minernodetestKey.txt
}


if [[ $# -eq 1 ]]; then
        startOneMinerNode $1
        exit 1
fi

startNodeNum=$1
while(( $startNodeNum<=$2 ))
do
        setCoinbase $startNodeNum
        let "startNodeNum++"
done

