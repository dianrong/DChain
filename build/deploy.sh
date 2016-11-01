#!/usr/bin/env bash

# @param 1: binary dirname
# @param 2: destination host login username
# @param 3: dest host ip
function setupEnv {
    local dir=$1
    local user=$2
    local ip=$3
    echo "setting up enviroment, 1. make ethereum home, 2, copy settings file, 3, copy genesis.txt"

    ssh $user'@'$ip 'mkdir -p ~/blockchain/eth/data/keystore'
    ssh $user'@'$ip 'mkdir -p ~/blockchain/common'
    ssh $user'@'$ip 'mkdir -p ~/blockchain/bin'
    ssh $user'@'$ip 'mkdir -p ~/blockchain/scripts'

    echo "dir is $dir"

    echo "copy genesis and config file to dst host"
    local dstConfigDir=`printf "%s@%s:~/blockchain/common" "$user" "$ip"`
    scp $dir/../common/properties.yaml $dstConfigDir
    scp $dir/deploy/genesis.txt $dstConfigDir

    echo "copy keystore to dst host"
    local dstEthKeyDir=`printf "%s@%s:~/blockchain/eth/data/keystore" "$user" "$ip"`
    scp $dir/deploy/keystore/* $dstEthKeyDir

    echo "copy scripts to dst host"
    local dstScriptsDir=`printf "%s@%s:~/blockchain/scripts" "$user" "$ip"`
    scp $dir/deploy/runGeth.sh $dstScriptsDir
    scp $dir/deploy/runCaServer.sh $dstScriptsDir
}

# $1: the dir of this file
# $2: target file to be copied
# $3: dest host username
# $4: dest host ip
function copyBin2DstHost {
    local dir=$1
    local exefile=$2 
    local user=$3
    local ip=$4
    echo "user is $user, ip is $ip"
    local targetURI=`printf "%s@%s:~/blockchain/bin" "$user" "$ip"`
    echo "copying executables $exefile to target place $targetURI"
    #ssh $user'@'$ip 'mkdir -p ~/blockchain'
    scp $dir/bin/$exefile $targetURI
}

function runBin {
    local user=$1
    local ip=$2
    local exe=$3
    local cmd=`printf "bash ~/blockchain/scripts/%s" "$exe"`
    echo "run $cmd"
    ssh $user'@'$ip $cmd
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


setupEnv `dirname $0` $1 $2
copyBin2DstHost `dirname $0` geth $1 $2
copyBin2DstHost `dirname $0` caserver $1 $2
runBin $1 $2 "runCaServer.sh"
runBin $1 $2 "runGeth.sh"
