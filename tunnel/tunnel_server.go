//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"net"
	"sync"
)

type TunnelServer struct {
	TcpServer
	wg      sync.WaitGroup
	newDoor func(*Tunnel) Service
}

func (self *TunnelServer) listen() {
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

func (self *TunnelServer) Start() error {
	err := self.buildListener()
	if err != nil {
		return err
	}

	self.wg.Add(1)
	go self.listen()
	return nil
}

func (self *TunnelServer) Reload() error {
	return nil
}

func (self *TunnelServer) handleClient(conn *net.TCPConn) {
	defer conn.Close()

	// try skip tgw
	err := skipTGW(conn)
	if err != nil {
		Error("skip tgw failed, source: %v", conn.RemoteAddr())
		return
	}

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
	tunnel := NewTunnel(conn)
	door := self.newDoor(tunnel)
	err = door.Start()
	if err != nil {
		Error("door start failed:%s", err.Error())
		return
	}
	door.Wait()
}

func (self *TunnelServer) Stop() {
	self.closeListener()
}

func (self *TunnelServer) Wait() {
	self.wg.Wait()
	Error("back door quit")
}

func NewTunnelServer(newDoor func(*Tunnel) Service) *TunnelServer {
	tunnelServer := new(TunnelServer)
	tunnelServer.TcpServer.addr = options.TunnelAddr
	tunnelServer.newDoor = newDoor
	return tunnelServer
}
