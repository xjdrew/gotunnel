## gotunnel
gotunnel is a secure tcp tunnel software. It uses persistent TCP connection(s) to communicate bettwen the client and the server, so it's not a port forwarder.

gotunnel could be added to any system using the TCP protocol. Make system structure evolve from
```
client <--------------> server
```
to
```
client <-> gotunnel <--------------> gotunnel <-> server
```
to gain gotunnel's valuable features, such as security and persistence. 

## Build

In your go workspace, run  the command below:
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

Some options:
* `secret`: for authentication and exchanging encryption key(s)
* `tunnels`: 0 means gotunnel will act as a server; Any value larger than 0 means gotunnel will act as a client, and build `tunnels` TCP connections to the server.
* `timeout`: if the packet body can't be read in `timeout` seconds, it will recreate this tunnel. It's useful if there is is a critical firewall between the gotunnel client and server.


## Example
Suppose you have a Squid server, and you use it as an HTTP proxy. Usually, you will start the server:
```
$ squid3 -a 8080
```
and use it on your local machine:
```
curl --proxy server:8080 http://example.com
```
It works fine but all traffic between your server and local machine is plaintext, so someone can moitor you easily. In this case, gotunnel could help to encrypt your traffic.

First, on your server, resart squid to listen on a local port, for example `127.0.0.1:3128`. Then start the gotunnel server listening on 8080 and use `127.0.0.1:3128` as the backend.
```
$ ./gotunnel -tunnels=0 -listen=:8001 -backend=127.0.0.1:3128 secret="your secret" -log=10 
```
Then, on your local machine, start  the gotunnel client:
```
$ ./gotunnel -tunnels=100 -listen="127.0.0.1:8080" -backend="server:8001" -secret="your secret" -log=10 
```

Then you can use `squid3` on you local port as before, but all of your traffic is encrypted. 

Besides that, you don't need to create and destory TCP connections between your local machine and the server, because gotunnel uses long-lived TCP connections as the low tunnel. In most cases, it would be faster.

## licence
The MIT License (MIT)

Copyright (c) 2015 xjdrew

