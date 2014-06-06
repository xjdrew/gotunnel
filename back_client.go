package main
import (
    "encoding/json"
    "io"
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
    upstream *Upstream
    linkset *LinkSet
    coor *Coor
    ch chan interface{}
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

func (self *BackClient) chooseHost() host *Host {
    upstream := self.upstream
    if upstream.weight <= 0 {
        return
    }
    v := rank.Intn(upstream.weight)
    for _, host := range upstream.Hosts {
        if host.Weight >= v {
            host = &host
            break
        }
        v -= host.Weight
    }
    return
}

func (self *BackClient) handleLink(linkid uint16, ch chan[]byte) {
    defer self.wg.Done()

    host := self.chooseHost()
    if host == nil {
        logger.Printf("choose host failed, linkid:%d", linkid)
        self.linkset.Reset(linkid)
        self.coor.SendLinkDestory(linkid)
        return
    }

    dest, err := net.DialTCP("tcp", nil, host.addr)
    if err != nil {
        logger.Printf("connect to host failed, linkid:%d, host:%d", linkid, host.Addr)
        self.linkset.Reset(linkid)
        self.coor.SendLinkDestory(linkid)
        return
    }

    link := NewLink(linkid, dest)

    self.wg.Add(1)
    go func() {
        defer self.wg.Done()
        link.Upload(self.coor)
    }

    err = link.Download(ch)
    if err != nil {
        self.linkset.Reset(linkid)
        self.coor.SendLinkDestory(linkid)
    }
}

func (self *BackClient) ctrl(cmd *CmdPayload) {
    linkid := cmd.Linkid
    switch cmd.Cmd {
        case LINK_CREATE:
            ch = make(chan[]byte, 256)
            err := self.linkset.Set(linkid, ch)
            if err != nil {
                logger.Printf("build link failed, linkid:%d, error:%s", linkid, err)
                self.coor.SendLinkDestory(linkid)
                return
            }
            self.wg.Add(1)
            go self.handleLink(linkid, ch)
        case LINK_DESTROY:
            ch, err := self.linkset.Reset(linkid)
            if err != nil {
                logger.Printf("close link failed, linkid:%d, error:%s", linkid, err)
                return
            }
            // close ch, don't write to ch again
            if ch != nil {
                close(ch)
            }
        default:
            logger.Printf("receive unknown cmd:%v", cmd)
    }
}

func (self *BackClient) data(payload *TunnelPayload) {
    linkid := payload.Linkid
    ch, err := self.LinkSet.Get(linkid)
    if err != nil {
        logger.Printf("illegal link, linkid:%d", linkid)
        return
    }

    if ch != nil {
        ch <- payload.Data
    else {
        logger.Printf("drop data becase no link, linkid:%d", linkid)
    }
}

func (self *BackClient) Dispatch() {
    defer self.wg.Done()
    for data := range self.ch {
        switch payload := data.(type) {
            case CmdPayload:
                self.ctrl(&payload)
            case TunnelPayload:
                self.data(payload)
            default:
                logger.Printf("unknown payload:%v", payload)
        }
    }
}

func (self *BackClient) Pump() {
    defer self.wg.Done()
    self.coor.Start()
    self.coor.Wait()
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

    self.ch := make(chan interface{}, 65535)
    self.coor := NewCoor(NewTunnel(conn), ch)

    self.wg.Add(1)
    go self.Pump()
    self.wg.Add(1)
    go self.Dispatch()
    return nil
}

func (self *BackClient) Stop() {
}

func (self *BackClient) Wait() {
    self.wg.Wait()
}

func NewBackClient(configFile string, backAddr string, capacity uint16) *BackClient {
    backClient := new(BackClient)
    backClient.configFile = configFile
    backClient.backAddr = backAddr
    backClient.linkset = NewLinkSet(capacity)
    return backClient
}

