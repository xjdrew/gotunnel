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
	TcpServer
	wg sync.WaitGroup
}

func (self *BackServer) listen() {
	defer self.wg.Done()

	for {
		conn, err := self.accept()
		if err != nil {
			Error("back server acceept failed:%s", err.Error())
			return
		}
		self.handleClient(conn)
	}
}

func (self *BackServer) Start() error {
	err := self.buildListener()
	if err != nil {
		return err
	}

	self.wg.Add(1)
	go self.listen()
	return nil
}

func (self *BackServer) handleClient(conn *net.TCPConn) {
	defer conn.Close()

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
	tunnel := NewTunnel(conn)
	frontDoor := NewFrontServer(tunnel)
	err := frontDoor.Start()
	if err != nil {
		Error("frontDoor start failed:%s", err.Error())
		return
	}
	frontDoor.Wait()
}

func (self *BackServer) Stop() {
	self.closeListener()
}

func (self *BackServer) Wait() {
	self.wg.Wait()
	Error("back door quit")
}

func NewBackServer() *BackServer {
	backDoor := new(BackServer)
	backDoor.TcpServer.addr = options.backAddr
	return backDoor
}
