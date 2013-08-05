#!/bin/bash 

REPO_DIR=$1

cd $REPO_DIR

status=`git status --porcelain`
diff=`git log origin/master..HEAD`

# statusx=$(git status --porcelain; echo sx)
# status="${statusx%sx}"

# diffx=$(git log origin/master..HEAD; echo dx)
# diff="${diffx%dx}"

if [ -n "$status" ] || [ -n "$diff" ] ; then
    echo '--------------------------------------------------------------------'
	echo "Checking status for $REPO_DIR"
	echo $status
	echo $diff
fi

# git diff --shortstat
