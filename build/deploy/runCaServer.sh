#!/usr/bin/env bash

ScriptsHome=`basedir $0`
echo "running casever in $ScriptsHome "
cd $ScriptsHome/..
screen -S "caserver" -L -d -m ./bin/caserver

