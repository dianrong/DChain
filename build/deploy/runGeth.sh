#!/usr/bin/env bash

ScriptsHome=`dirname $0`
EthereumHome=$ScriptsHome/../eth
echo "assume geth is placed under ~/blockchain/bin, Ethereum home is  $EthereumHome"
cd $ScriptsHome/..
if [ ! -f $EthereumHome/data/nodekey ];then
    echo "init with genesis"
    ./bin/geth --datadir $EthereumHome/data init ./common/genesis.txt
fi
./bin/geth --datadir $EthereumHome/data --verbosity 6 --networkid 8 --ipcapi "admin,db,eth,debug,miner,net,shh,txpool,personal,web3" --port 62360 --rpcaddr 0.0.0.0 --rpcapi "web3,eth,miner,txpool,personal,admin,db,shh,debug,net" --rpc --rpcport 62361 --natspec --nodiscover 2>$EthereumHome/eth.log

