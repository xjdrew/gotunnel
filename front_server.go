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
	wg   sync.WaitGroup
	coor *Coor
}

func (self *FrontServer) pump() {
	defer self.wg.Done()
	self.coor.Start()
	self.coor.Wait()
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

	self.wg.Add(1)
	go self.pump()
	return nil
}

func (self *FrontServer) handleClient(conn *net.TCPConn) {
	defer self.wg.Done()

	linkid := self.coor.AcquireId()
	if linkid == 0 {
		Error("alloc linkid failed, source: %v", conn.RemoteAddr())
		conn.Close()
		return
	}

	ch := make(chan []byte, 256)
	err := self.coor.Set(linkid, ch)
	if err != nil {
		//impossible
		conn.Close()
		Error("set link failed, linkid:%d, source: %v", linkid, conn.RemoteAddr())
		return
	}

	Info("new link:%d, source: %v", linkid, conn.RemoteAddr())
	defer self.coor.ReleaseId(linkid)

	self.coor.SendLinkCreate(linkid)

	link := NewLink(linkid, conn)
	err = link.Pump(self.coor, ch)
	if err != nil {
		self.coor.Reset(linkid)
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
	frontServer.coor = NewCoor(tunnel, nil)
	frontServer.TcpServer.addr = options.frontAddr
	return frontServer
}
