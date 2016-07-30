[![Build Status](https://travis-ci.org/xjdrew/gotunnel.svg?branch=master)](https://travis-ci.org/xjdrew/gotunnel)

## gotunnel
gotunnel is a secure tcp tunnel software. It can use tcp or udp connectioin as low level tunnel.

gotunnel could be added to any c/s system using tcp protocal. Make system structure evolve from
```
client <--------------> server
```
to
```
client <-> gotunnel <--------------> gotunnel <-> server
```
to gain gotunnel's valuable features, such as secure and persistent. 

## build

###1. download codebase
```bash
go get -u -d github.com/xjdrew/gotunnel
```
###2. build udt
```bash
cd ${GOPATH}/src/github.com/xjdrew/go-udtwrapper/udt4/src && make libudt.a && cp libudt.a ${GOPATH}
```
###3. build gotunnel
```bash
GOPATH=${GOPATH} CGO_LDFLAGS=-L${GOPATH} go install github.com/xjdrew/gotunnel
```

* build automatically

You can run the script [install.sh](https://github.com/xjdrew/gotunnel/blob/master/install.sh) directly:
```
bash <<(curl -fsSL https://github.com/xjdrew/gotunnel/blob/master/install.sh)
```

## Usage

```
usage: bin/gotunnel
  -backend string
        backend address (default "127.0.0.1:1234")
  -listen string
        listen address (default ":8001")
  -log uint
        log level (default 1)
  -secret string
        tunnel secret (default "the answer to life, the universe and everything")
  -timeout int
        tunnel read/write timeout (default 3)
  -tunnels uint
        low level tunnel count, 0 if work as server
  -udt
        udt tunnel
```

some options:
* secret: for authentication and exchanging encryption key
* tunnels: 0 means gotunnel will and as server; Any value larger than 0 means gotunnel will work as client, and build *tunnels* tcp connections to server.
* timeout: if can't read a packet body in *timeout* seconds, will recreate this tunnel. It's useful if theres is a critical firewall between gotunnel client and server.


## Example
Suppose you have a squid server, and you use it as a http proxy. Usually, you will start the server:
```
$ squid3 -a 8080
```
and use it on your pc:
```
curl --proxy server:8080 http://example.com
```
It works fine but all traffic between your server and pc is plaintext, so someone can monitor you easily. In this case, gotunnel could help to encrypt your traffic.

First, on your server, resart squid to listen on a local port, for example **127.0.0.1:3128**. Then start gotunnel server listen on 8080 and use **127.0.0.1:3128** as backend.
```
$ ./gotunnel -listen=:8001 -backend=127.0.0.1:3128 secret="your secret" -log=10 
```
Second, on your pc, start gotunnel client:
```
$ ./gotunnel -tunnels=100 -listen="127.0.0.1:8080" -backend="server:8001" -secret="your secret" -log=10 
```

Then you can use squid3 on you local port as before, but all your traffic is encrypted. 

Besides that, you don't need to create and destory tcp connection between your pc and server, because gotunnel use long-live tcp connections as low tunnel. In most cases, it would be faster.

## licence
The MIT License (MIT)

Copyright (c) 2015 xjdrew

