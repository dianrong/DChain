#!/usr/bin/env bash

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


# $1: base dir
function setupLocalEnv() {
    local dir=$1
    local BLOCKCHAINDIR=$2
    for tdir in bin scripts common eth;do
        echo "creating dir $BLOCKCHAINDIR/$tdir"
        mkdir -p $BLOCKCHAINDIR/$tdir
    done

    mkdir -p $BLOCKCHAINDIR/eth/data/keystore

    echo "copy geth and caserver"
    cp $dir/bin/geth $BLOCKCHAINDIR/bin/
    cp $dir/bin/caserver  $BLOCKCHAINDIR/bin/

    echo "copy keystore files"
    scp $dir/deploy/keystore/* $BLOCKCHAINDIR/eth/data/keystore/

    echo "copy ca configuration file"
    cp $dir/../common/properties.yaml $BLOCKCHAINDIR/common/
    cp $dir/deploy/genesis.txt $BLOCKCHAINDIR/common/

    echo "copying scripts"
    cp $dir/deploy/runGeth.sh $BLOCKCHAINDIR/scripts/
    cp $dir/deploy/runCaServer.sh $BLOCKCHAINDIR/scripts/
}


# $1 target dir
# $2 username
# $3 ip
function copy2Remote() {
    local dir=$1
    local user=$2
    local ip=$3

    scp -rp $dir $user@$ip:~/
}

# $1 the dir to be archived
# $2 the archived file name
function createTar() {
    local dir=$1
    local tarName=$2
    cd $dir/..
    tar czf $tarName blockchain
    cd -
    mv $dir/../$tarName ./
}


if [ $# -lt 2 ];then
    echo "Wrong usage, This script take 2 arguments"
    Usage
    exit 1
else
    Usage
fi

currentDate=`date +"%Y%m%d_%H%M%S"`
currentDir=`dirname $0`
blockchainDir=$currentDir/deploy/tmp/$currentDate/blockchain
echo "current dir is $currentDir, bc dir is $blockchainDir"
setupLocalEnv $currentDir $blockchainDir
createTar $blockchainDir blockchain_$currentDate_`uname`.tar.gz
copy2Remote $blockchainDir $1 $2

echo "Finished setting up local and remote enviroment"


runBin $1 $2 "runCaServer.sh"

runBin $1 $2 "runGeth.sh"
