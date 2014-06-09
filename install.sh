#!/bin/bash

mkdir -p gospace/src
export GOPATH=`pwd`/gospace
git clone https://github.com/xjdrew/gotunnel gospace/src/gotunnel
go install gotunnel

# finish
echo "gotunnel is in gospace/bin/, go and run"

