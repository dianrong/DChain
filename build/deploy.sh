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
    echo "Usage : $0 {ca|geth|both} username ipadddress"
    echo "This scripts is used to deploy Blockchain app into target machine, and run it"
    echo "Before using it, make sure the target machine can be logged on via ssh without password"
}


# $1: base dir
function setupLocalEnv() {
    local dir=$1
    local BLOCKCHAINDIR=$2
    local mode=$3
    for tdir in bin scripts common eth;do
        echo "creating dir $BLOCKCHAINDIR/$tdir"
        mkdir -p $BLOCKCHAINDIR/$tdir
    done

    echo "--->STEP1: Setting up local envrioment"
    mkdir -p $BLOCKCHAINDIR/eth/data/keystore

    echo "...copy geth and/or caserver"
    case $mode in
      ca)
        cp $dir/bin/caserver $BLOCKCHAINDIR/bin/
        cp $dir/deploy/runCaServer.sh $BLOCKCHAINDIR/scripts/
        cp $dir/deploy/caserver.service $BLOCKCHAINDIR/scripts/caserver.service
        ;;
      geth)
        cp $dir/bin/geth $BLOCKCHAINDIR/bin/
        cp $dir/deploy/runGeth.sh $BLOCKCHAINDIR/scripts/
        cp $dir/deploy/geth.service $BLOCKCHAINDIR/scripts/geth.service
        ;;
      *)
        cp $dir/bin/caserver $BLOCKCHAINDIR/bin/
        cp $dir/deploy/runCaServer.sh $BLOCKCHAINDIR/scripts/
        cp $dir/deploy/caserver.service $BLOCKCHAINDIR/scripts/caserver.service
        cp $dir/bin/geth $BLOCKCHAINDIR/bin/
        cp $dir/deploy/runGeth.sh $BLOCKCHAINDIR/scripts/
        cp $dir/deploy/geth.service $BLOCKCHAINDIR/scripts/geth.service
        ;;
    esac

    echo "...copy keystore files"
    scp $dir/deploy/keystore/* $BLOCKCHAINDIR/eth/data/keystore/

    echo "...copy ca configuration file"
    cp $dir/../common/properties.yaml $BLOCKCHAINDIR/common/
    cp $dir/deploy/genesis.txt $BLOCKCHAINDIR/common/

    echo "...copy install script"
    cp $dir/deploy/install.sh $BLOCKCHAINDIR/scripts/install.sh
    echo "<---Finished setting up local envrioment"
}


# $1 target dir
# $2 username
# $3 ip
function copy2Remote() {
    local dir=$1
    local user=$2
    local ip=$3

    echo "---> STEP: Setting up remote envrioment"
    scp -rp $dir $user@$ip:~/
    echo "<--- Finished Setting up remote enviroment"
}

# $1 the dir to be archived
# $2 the archived file name
function createTar() {
    local dir=$1
    local tarName=$2
    echo "---> STEP: Creating archived file"
    echo "tarname is $tarName"
    cd $dir/..
    tar czf $tarName blockchain*
    cd -
    mv $dir/../$tarName ./
    echo "<--- STEP: Finished achiving file"
}

if [ $# -ge 1 ];then
    if [ $1 == "-h" ];then
        Usage
        exit 0
    fi
fi

currentDate=`date +"%Y%m%d_%H%M%S"`
currentDir=`dirname $0`
blockchainDir=$currentDir/deploy/tmp/$currentDate/blockchain
echo "current dir is $currentDir, bc dir is $blockchainDir"

mode=both
case $1 in
  ca)
    mode=ca
    ;;
  geth)
    mode=geth
    ;;
esac

setupLocalEnv $currentDir $blockchainDir $mode
createTar $blockchainDir blockchain_${currentDate}_`uname`.tar.gz

if [ $# -lt 3 ];then
    echo "If you need to deploy this to a remote host, two extra argument required"
    exit 1
fi

copy2Remote $blockchainDir $2 $3

runBin $1 $2 "runCaServer.sh"

runBin $1 $2 "runGeth.sh"

