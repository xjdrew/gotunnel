## gotunnel
### features
*  support rc4 encryption
*  support tgw header
*  support reload back end services

### tunnel stack

SOURCE   | DESTINATION
:--------|------------:
TcpConn  | TCPConn
Link     |    Link
Coor     |    Coor
Tunnel   |  Tunnel

### 组件
* back server, back client: 负责建立tunnel
* front server: 负责接受玩家连接
* coor, link set: 负责管理link
* link: 抽象的概念，实际连接在tunnel上的映射

## build
如果没有搭建过go 的workspace，参考install.sh里面的脚本

go install gotunnel


## run

###launch gate(defaut port 8001, 8002):
```
./gotunnel -gate
```

###launch node(defaut service port 1234):
```
./gotunnel -back_addr=127.0.0.1:8002 settings.conf
```

###launch service:
```
nc -lk 1234
```

###launch client:
```
nc 127.0.0.1 8001
```

## benchmark
### 测试环境
* OS：Ubuntu quantal (12.10)
* CPU: Intel(R) Core(TM) i5-2400 CPU @ 3.10GHz
* golang:  go1.2.1 linux/amd64

### gotunnel启动参数
* gate:./gotunnel -gate
* node:./gotunnel -back_addr 127.0.0.1:8002 settings.conf

### nginx 作为后端
* nginx/1.2.1
* 测试工具使用ab (version 2.3)， 10w次请求
* 下表对比（Requests per second）

并发数      |    no tunnel  |    go tunnel(CPU=1) | go tunnel(CPU=2)
:-----------|:--------------|:--------------------|:-------------------
1           |    7473       |    1902             | 1826
10          |    21734      |    7508             | 7627
100         |    20519      |    8729             | 10437
500         |    18548      |    8968             | 10287
1000        |    18255      |    9141             | 10127

* 并发上去之后，go的效率提升不明显，在100并发数附近结果最好，有点奇怪。
* 也使用redis-benchmark进行了测试，结果略好一点。





