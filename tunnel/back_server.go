//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

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
		Debug("back server, new connection from %v", conn.RemoteAddr())
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

func (self *BackServer) Reload() error {
	return nil
}

func (self *BackServer) handleClient(conn *net.TCPConn) {
	defer conn.Close()

	// try skip tgw
	err := skipTGW(conn)
	if err != nil {
		Error("skip tgw failed, source: %v", conn.RemoteAddr())
		return
	}

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
	tunnel := NewTunnel(conn)
	frontDoor := NewFrontServer(tunnel)
	err = frontDoor.Start()
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
	backDoor.TcpServer.addr = options.BackAddr
	return backDoor
}
