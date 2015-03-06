## gotunnel
gotunnel在客户端和服务器之间建立一条tcp连接作为隧道，客户端和服务器把所有应用连接上的请求回应打包成字节流加密后通过隧道传输到对端，并把字节流还原到相应的连接上去，使隧道对应用层透明。

通过使用gotunnel，可以减少客户端和服务器之间频繁的连接建立和断开，提升应用的效率，尤其是他们之间隔着防火墙的时候。

### tunnel stack

SOURCE   | DESTINATION
:--------|------------:
TcpConn  | TCPConn
Link     | Link
Hub      | Hub
Tunnel   | Tunnel

## build
如果没有搭建过go 的workspace，参考install.sh里面的脚本

go install gotunnel

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

