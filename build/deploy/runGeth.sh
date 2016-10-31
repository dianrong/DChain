#!/usr/bin/env bash

EthereumHome=$HOME/blockchain/eth
echo "assume geth is placed under ~/blockchain/bin, Ethereum home is  $EthereumHome"
cd ~/blockchain
if [ -f $EthereumHome/data/nodekey ];then
    echo "init with genesis"
    geth --datadir $EthereumHome/data init ./common/genesis.txt
fi
screen -d -m ./bin/geth --datadir $EthereumHome/data --verbosity 6 --networkid 7 --ipcapi "admin,db,eth,debug,miner,net,shh,txpool,personal,web3" --port 62360 --rpcaddr 0.0.0.0 --rpcapi "web3,eth,miner,txpool,personal,admin,db,shh,debug,net" --rpc --rpcport 62361 --natspec --nodiscover console 2>$EthereumHome/eth.log

