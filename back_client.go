package main
import (
    "encoding/json"
    "net"
    "os"
    "sync"
)

type Host struct {
    Addr string
    Weight int

    addr *net.TCPAddr
}

type Upstream struct {
    Hosts []Host
    weight int 
}

type BackClient struct {
    configFile string
    backAddr string
    wg sync.WaitGroup
    coor *Coor
    upstream *Upstream
}

func (self *BackClient) readSettings() (upstream *Upstream, err error) {
    fp, err := os.Open(self.configFile)
    if err != nil {
        logger.Printf("open config file failed:%s", err.Error())
        return
    }
    defer fp.Close()
    
    upstream = new(Upstream)
    dec := json.NewDecoder(fp)
    err = dec.Decode(upstream)
    if err != nil {
        logger.Printf("decode config file failed:%s", err.Error())
        return 
    }

    for i := range upstream.Hosts {
        host := &upstream.Hosts[i]
        host.addr, err = net.ResolveTCPAddr("tcp", host.Addr)
        if err != nil {
            logger.Printf("resolve local addr failed:%s", err.Error())
            return
        }
        upstream.weight += host.Weight
    }

    logger.Printf("config:%v", upstream)
    return 
}

func (self *BackClient) Start() error {
    upstream, err := self.readSettings()
    if err != nil {
        return err
    }
    self.upstream = upstream
    
    addr, err := net.ResolveTCPAddr("tcp", self.backAddr)
    if err != nil {
        return err
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return err
    }

    ch := make(chan interface{}, 65535)
    self.wg.Add(1)
    go func() {
        defer self.wg.Done()
        defer conn.Close()

        tunnel := NewTunnel(conn)
        self.coor.SetTunnel(tunnel, ch)
        self.coor.Start()
        self.coor.Wait()
    }()

    self.wg.Add(1)
    go func() {
        defer self.wg.Done()
        for payload := range ch {
            logger.Printf("payload:%v", payload)
        }
    }()

    return nil
}

func (self *BackClient) Stop() {
}

func (self *BackClient) Wait() {
    self.wg.Wait()
}

func NewBackClient(configFile string, backAddr string, coor *Coor) *BackClient {
    var wg sync.WaitGroup
    return &BackClient{configFile, backAddr, wg, coor, nil}
}

