#!/usr/bin/env bash

function setupEnv {
    local dir=$1
    local user=$2
    local ip=$3
    echo "setting up enviroment, 1. make ethereum home, 2, copy settings file, 3, copy genesis.txt"

    ssh $user'@'$ip 'mkdir -p ~/blockchain/eth/data/keystore'
    ssh $user'@'$ip 'mkdir -p ~/blockchain/common'
    ssh $user'@'$ip 'mkdir -p ~/blockchain/bin'

    echo "dir is $dir"

    local targetConfigDir=`printf "%s@%s:~/blockchain/common" "$user" "$ip"`
    scp $dir/../common/properties.yaml $targetConfigDir
    scp $dir/deploy/genesis.txt $targetConfigDir

    local targetEthKeyDir=`printf "%s@%s:~/blockchain/eth/data/keystore" "$user" "$ip"`
    scp $dir/deploy/keystore/* $targetEthKeyDir
}

# $1 is the ip address
function copy2Machine {
    local user=$1
    local ip=$2
    echo "user is $user, ip is $ip"
    local targetURI=`printf "%s@%s:~/blockchain/bin" "$user" "$ip"`
    echo "copying executables to target place $targetURI"
    #ssh $user'@'$ip 'mkdir -p ~/blockchain'
    scp ./bin/geth $targetURI
}

function runGeth {
    local user=$1
    local ip=$2
    ssh $user'@'$ip 'bash -s' < ./deploy/runGeth.sh
}


function Usage {
    echo "Usage : $0 username ipadddress"
    echo "This scripts is used to deploy Blockchain app into target machine, and run it"
    echo "Before using it, make sure the target machine can be logged on via ssh without password"
}

if [ $# -lt 2 ];then
    echo "Wrong usage, This script take 2 arguments"
    Usage
    exit 1
else
    Usage
fi

#setupEnv `dirname $0` $1 $2
#copy2Machine $1 $2
runGeth $1 $2
