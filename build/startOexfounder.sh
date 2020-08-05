# 此脚本用于启动测试网创世节点
openHttp="--http_host 0.0.0.0 --http_port 8080"

nohup ./oex --genesis=../testnet.json --datadir=./data/founder --miner_start --contractlog --p2p_listenaddr :9090 $openHttp --http_modules=fee,miner,dpos,account,txpool,oex >> founder.log &
sleep 5s
./oex miner -i ./data/founder/oex.ipc setcoinbase "oexchain.founder" keys/founderKey.txt
