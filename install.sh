#!/bin/bash

mkdir -p gospace/src
cd gospace
export GOPATH=`pwd`

go get github.com/xjdrew/gotunnel
go install github.com/xjdrew/gotunnel

# finish
echo "gotunnel is in gospace/bin/, go and run"

