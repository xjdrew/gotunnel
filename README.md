## gotunnel
gotunnel在客户端和服务器之间建立一条tcp连接作为隧道，客户端和服务器把所有连接上的请求和相应打包成字节流加密后通过隧道传输到对端，然后把数据还原到相应的连接上去。通过使用gotunnel，可以减少客户端和服务器之间频繁的连接建立和断开，提升应用的效率，尤其是他们之间隔着防火墙的时候。

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
  -rc4="the answer to life, the universe and everything": rc4 key, disable if no key
  -server="": server address, empty if work as server
```

#### launch tunnel server

```
./gotunnel setting.conf
```

#### launch tunnel client
```
./gotunnel -server="127.0.0.1:8001" -listen=":8002"
```

