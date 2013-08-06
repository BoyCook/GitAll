#!/bin/bash 

PACKAGE_DIR=/usr/local/lib/node_modules/gitclone

if [ -d $PACKAGE_DIR ]
then
    echo "Package in installed at $PACKAGE_DIR - deleting"
    rm -rf $PACKAGE_DIR 
fi	

npm install . -g
