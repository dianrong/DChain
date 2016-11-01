#!/usr/bin/env bash

ScriptsHome=$HOME/blockchain/scripts
echo "running casever in $ScriptsHome "
cd $ScriptsHome/..
screen -S "caserver" -L -d -m ./bin/caserver

