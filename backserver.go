//
//   date  : 2014-06-05
//   author: xjdrew
//

package main

import (
    "net"
    "sync"
)

type BackServer struct {
    *TcpServer
    coor *Coor
    wg sync.WaitGroup
}

func (self *BackServer) Start() error {
    err := self.buildListener()
    if err != nil {
        return err
    }
    
    self.wg.Add(1)
    go func() {
        defer self.wg.Done()
        conn, err := self.accept()
        if err != nil {
            return
        }
        self.closeListener()
        self.handleClient(conn)
    }()
    return nil
}

func (self *BackServer) handleClient(conn *net.TCPConn) {
    defer self.wg.Done()
    defer conn.Close()

    logger.Printf("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
    tunnel := NewTunnel(conn)
    self.coor.SetTunnel(tunnel)
    self.coor.Start()
    self.coor.Wait()
}

func (self *BackServer) Stop() {
    self.closeListener()
}

func (self *BackServer) Wait() {
    self.wg.Wait()
}

func NewBackServer(addr string, coor *Coor) *BackServer {
    var wg sync.WaitGroup
    return &BackServer{NewTcpServer(addr), coor, wg}
}

