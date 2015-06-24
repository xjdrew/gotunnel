## gotunnel
gotunnel is a secure tcp tunnel software. It use persistent tcp connection(s) to comminicate bettwen client and server, so it's not a port forwarder.

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

In your go workspace, run command as below:
```bash
go get -u git@github.com:xjdrew/gotunnel.git
```
If you don't known how to create a golang workspace, please see [install.sh](https://github.com/xjdrew/gotunnel/blob/master/install.sh)

## Usage

```
usage: bin/gotunnel
  -backend="127.0.0.1:1234": backend address
  -listen=":8001": listen address
  -log=1: log level
  -secret="the answer to life, the universe and everything": tunnel secret
  -timeout=10: tunnel read/write timeout
  -tunnels=1: low level tunnel count, 0 if work as server
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
It works fine but all traffic between your server and pc is plaintext, so someone can moitor you easily. In this case, gotunnel could help to encrypt your traffic.

First, on your server, resart squid to listen on a local port, for example **127.0.0.1:3128**. Then start gotunnel server listen on 8080 and use **127.0.0.1:3128** as backend.
```
$ ./gotunnel -tunnels=0 -listen=:8001 -backend=127.0.0.1:3128 secret="your secret" -log=10 
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

