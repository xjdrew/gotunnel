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
    *TcpServer
    coor *Coor
    wg sync.WaitGroup
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

    link := NewLink(linkid, conn)
    self.coor.SendLinkCreate(linkid)

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

func (self *FrontServer) Stop() {
    self.closeListener()
}

func (self *FrontServer) Wait() {
    self.wg.Wait()
}

func NewFrontServer(addr string, coor *Coor, capacity uint16) *FrontServer {
    frontServer := new(FrontServer)
    frontServer.TcpServer.addr = addr
    frontServer.coor = coor
    frontServer.linkset = NewLinkSet(capacity)
    return frontServer
}

