#!/bin/bash 

ROOT_DIR=$1
REPO=$2
TARGET_DIR=$3

cd $ROOT_DIR
git clone $REPO $TARGET_DIR
