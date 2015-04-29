## gotunnel
gotunnel是一个加密tcp隧道, 可以透明的加到任何使用tcp协议的应用和服务器之间，提供认证、加密和高效数据传输的能力。

根据启动参数的不同，gotunnel可以客户端或者服务器端的模式工作。

客户端启动的时候会同服务器端建立*tunnel_count*条tcp连接，连接使用*secret*互相认证，并交换随机数作为后续的加密密钥。如果在运行过程中，连接断开，gotunnel会自动重连。

当客户端收到应用发来的连接请求时，会同服务器协商，建立一条虚拟的link；服务器同应用服务器之间建立一条tcp连接，并同虚拟link关联起来。应用客户端发送的任何数据包都会原封不动的传送到应用服务器，gotunnel对应用完全透明。

通过使用gotunnel，可以减少客户端和服务器之间频繁的连接建立和断开，提升应用的效率，尤其是他们之间隔着防火墙的时候。

## install
```bash
go get -u git@github.com:xjdrew/gotunnel.git
```
如果没有搭建过go 的workspace，请参考install.sh里面的脚本

## run
```
$ bin/gotunnel 
usage: bin/gotunnel [configFile]
  -listen=":8001": host:port gotunnel listen on
  -log=1: larger value for detail log
  -secret="the answer to life, the universe and everything": connection secret, disable if has none
  -server="": server address, empty if work as server
  -tunnel_count=1: underlayer tunnel count
```

## useage
```bash
$ source env.sh
$ go build github.com/xjdrew/gotunnel
$ go build src/github.com/xjdrew/gotunnel/tests/sender.go
$ go build src/github.com/xjdrew/gotunnel/tests/receiver.go

# launch tunnel server
$ ./gotunnel -log=10 src/github.com/xjdrew/gotunnel/tests/test.conf

# launch tunnel client
$ ./gotunnel -log=10 -listen=":8003" -server="127.0.0.1:8001"

# launch receiver
$ ./receiver

# lauch sender
$ ./sender -remote=127.0.0.1:8003
```

