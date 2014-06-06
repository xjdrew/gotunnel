//
//   date  : 2014-06-05
//   author: xjdrew
//

package main

import (
    "net"
    "sync"
)

type FrontServer struct {
    TcpServer
    tunnel *Tunnel
    linkset *LinkSet

    wg sync.WaitGroup
    coor *Coor
    tunnelCh chan interface{}
}

func (self *FrontServer) pump() {
    defer self.wg.Done()
    self.coor.Start()
    self.coor.Wait()
}

func (self *FrontServer) ctrl(cmd *CmdPayload) {
    linkid := cmd.Linkid
    switch cmd.Cmd {
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

func (self *FrontServer) data(payload *TunnelPayload) {
    linkid := payload.Linkid
    ch, err := self.linkset.Get(linkid)
    if err != nil {
        logger.Printf("illegal link, linkid:%d", linkid)
        return
    }

    if ch != nil {
        ch <- payload.Data
    } else {
        logger.Printf("drop data becase no link, linkid:%d", linkid)
    }
}

func (self *FrontServer) dispatch() {
    defer self.wg.Done()
    for payload := range self.tunnelCh {
        switch payload := payload.(type) {
            case *CmdPayload:
                self.ctrl(payload)
            case *TunnelPayload:
                self.data(payload)
            default:
                logger.Printf("unknown payload type:%T", payload)
        }
    }
}

func (self *FrontServer) Start() error {
    err := self.buildListener()
    if err != nil {
        return err
    }

    self.wg.Add(1)
    go func() {
        defer self.wg.Done()
        for {
            conn, err := self.accept()
            if err != nil {
                break
            }
            self.wg.Add(1)
            go self.handleClient(conn)
        }
    }()

    self.tunnelCh = make(chan interface{}, 65535)
    self.coor = NewCoor(self.tunnel, self.tunnelCh)

    self.wg.Add(1)
    go self.pump()
    self.wg.Add(1)
    go self.dispatch()
    return nil
}

func (self *FrontServer) handleClient(conn *net.TCPConn) {
    defer self.wg.Done()

    linkid := self.linkset.AcquireId()
    if linkid == 0 {
        logger.Printf("alloc linkid failed, source: %v", conn.RemoteAddr())
        conn.Close()
        return
    }

    ch := make(chan []byte, 256)
    err := self.linkset.Set(linkid, ch)
    if err != nil {
        //impossible
        conn.Close()
        logger.Panicf("set link failed, linkid:%d, source: %v", linkid, conn.RemoteAddr())
        return
    }

    logger.Printf("new link:%d, source: %v", linkid, conn.RemoteAddr())
    defer self.linkset.ReleaseId(linkid)

    self.coor.SendLinkCreate(linkid)

    link := NewLink(linkid, conn)
    err = link.Pump(self.coor, ch)
    if err != nil {
        self.linkset.Reset(linkid)
    }
}

func (self *FrontServer) Stop() {
    self.closeListener()
}

func (self *FrontServer) Wait() {
    self.wg.Wait()
}

func NewFrontServer(tunnel *Tunnel) *FrontServer {
    frontServer := new(FrontServer)
    frontServer.tunnel = tunnel
    frontServer.TcpServer.addr = options.frontAddr
    frontServer.linkset = NewLinkSet(options.capacity)
    return frontServer
}

