## 原理
### 数据流
conn <-> link <-> (link set) <-> coor <-> tunnel ======== tunnel <-> coor <-> (link set) <-> link <-> conn

### 组件
*back server, back client: 负责建立tunnel
*front server: 负责接受玩家连接
*coor, link set: 负责管理link
*link: 抽象的概念，实际连接在tunnel上的映射

## make
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

