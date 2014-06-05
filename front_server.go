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

    
    wch := make(chan []byte, 256)
    linkid := self.coor.AddLink(wch)
    if linkid == 0 {
        logger.Printf("alloc linkid failed, source: %v", conn.RemoteAddr())
        conn.Close()
        return
    }

    logger.Printf("new link:%d, source: %v", linkid, conn.RemoteAddr())
    defer self.coor.RemLink(linkid)

    self.coor.Send(LINK_CREATE, linkid, nil)
    defer self.coor.Send(LINK_DESTROY, linkid, nil)

    go func() {
        for buffer := range wch {
            c := 0
            for c < len(buffer) {
                n, err := conn.Write(buffer[c:])
                if err != nil {
                    break
                }
                c += n
            }
        }
        // passive disconnected
        conn.Close()
    }()

    for {
        buffer := make([]byte, 0xff)
        n, err := conn.Read(buffer)
        if err != nil {
            break
        }
        self.coor.Send(LINK_DATA, linkid, buffer[:n])
    }
}

func (self *FrontServer) Stop() {
    self.closeListener()
}

func (self *FrontServer) Wait() {
    self.wg.Wait()
}

func NewFrontServer(addr string, coor *Coor) *FrontServer {
    var wg sync.WaitGroup
    return &FrontServer{NewTcpServer(addr), coor, wg}
}

