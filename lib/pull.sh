#!/bin/bash 

REPO_DIR=$1
SSH_URL=$2
HTTPS_URL=$3

cd $REPO_DIR

REMOTE_URL=`git config --get remote.origin.url`

if [[ "$REMOTE_URL" == *$SSH_URL* ]] || [[ "$REMOTE_URL" == *$HTTPS_URL* ]]
then
	git pull
else 
	echo "Ignoring non users repository [$REMOTE_URL] @ [$REPO_DIR]";
fi
