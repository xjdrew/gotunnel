## 原理
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

