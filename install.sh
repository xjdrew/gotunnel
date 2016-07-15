#!/bin/bash

mkdir -p gospace
export GOPATH=`pwd`/gospace

go get -u -d github.com/xjdrew/gotunnel
cd ${GOPATH}/src/github.com/xjdrew/go-udtwrapper/udt4/src && make libudt.a && cp libudt.a ${GOPATH}
CGO_LDFLAGS=-L${GOPATH} go install github.com/xjdrew/gotunnel

# finish
echo "gotunnel is in gospace/bin/, go and run"

