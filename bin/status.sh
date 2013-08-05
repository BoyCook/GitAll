#!/bin/bash 

REPO_DIR=$1

echo '--------------------------------------------------------------------'
echo "Checking status for $REPO_DIR"
cd $REPO_DIR
git status --porcelain
git log origin..

# git diff --shortstat
