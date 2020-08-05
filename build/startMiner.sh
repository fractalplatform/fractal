# 脚本用法：
# ./startMiner.sh 1 10 : 启动矿工1~矿工10
# 注意：
# 1：当创世文件中配置的节点无法连接时（譬如超过连接总数），需要通过setP2PNode.sh来配置新的节点
# 2: 此处未启动挖矿，如需启动挖矿，需要使用setCoinbase.sh脚本
#start miner nodes
#

declare -a p2pNodes

index=0
for line in `cat filename(p2pNodes.txt)`
do
	p2pNodes[$index]=line
	let "index++"
done
echo ${p2pNodes[*]}


if [[ $# -eq 0 ]]; then
	echo "command=>$0, no parameters"
	exit 1
fi

function startOneMinerNode () 
{
	minerName="minernodetest$1"
	p2pPort=`expr 8090 + $1`
	httpPort=`expr 8900 + $1 + $1`
	wsPort=`expr $httpPort + 1`
	echo "minerName=$minerName, p2pport=$p2pPort, httpPort=$httpPort, wsPort=$wsPort"
	#mkdir ./data/$minerName
	nohup ./oex --genesis=../testnet.json --datadir=./data/$minerName --contractlog --p2p_listenaddr :$p2pPort --http_port $httpPort --ws_port $wsPort --http_modules=fee,miner,dpos,account,txpool,oex >> logs/$minerName.log &
	#sleep 5
	#./oex miner -i ./data/$minerName/oex.ipc setcoinbase "$minerName" keys/minernodetestKey.txt
}


if [[ $# -eq 1 ]]; then
	startOneMinerNode $1
	exit 1
fi

startNodeNum=$1
while(( $startNodeNum<=$2 ))
do
	startOneMinerNode $startNodeNum
	let "startNodeNum++"	
done
