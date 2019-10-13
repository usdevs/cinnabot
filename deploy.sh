#!/bin/bash

export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
CINNABOT=$GOPATH/src/github.com/usdevs/cinnabot
cd $CINNABOT

# pull from github
git pull origin master
COMMITHEAD=$(git rev-parse --short HEAD)
LASTUPDATED=$(git log -1 --format=%cr)

# build binary
cd main
go build

# run binary
./main