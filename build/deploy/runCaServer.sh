#!/usr/bin/env bash

ScriptsHome=`dirname $0`
echo "running casever in $ScriptsHome "
cd $ScriptsHome/..
screen -S "caserver" -L -d -m ./bin/caserver

