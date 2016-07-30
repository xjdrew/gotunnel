get-deps:
	go get -u -d github.com/xjdrew/go-udtwrapper
	cd ${GOPATH}/src/github.com/xjdrew/go-udtwrapper/udt4/src && make libudt.a && cp libudt.a ${GOPATH}
	GOPATH=${GOPATH} CGO_LDFLAGS=-L${GOPATH} go install github.com/xjdrew/gotunnel
