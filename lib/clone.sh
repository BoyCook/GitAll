#!/bin/bash 

ROOT_DIR=$1
REPO=$2
TARGET=$3

cd $ROOT_DIR
git clone $REPO $TARGET
