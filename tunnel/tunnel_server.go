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
	hubs map[*ServerHub]bool
	wg   sync.WaitGroup
	rw   sync.RWMutex
}

func (self *TunnelServer) addHub(hub *ServerHub) {
	self.rw.Lock()
	self.hubs[hub] = true
	self.rw.Unlock()
}

func (self *TunnelServer) removeHub(hub *ServerHub) {
	self.rw.Lock()
	delete(self.hubs, hub)
	self.rw.Unlock()
}

func (self *TunnelServer) handleClient(conn *net.TCPConn) {
	defer self.wg.Done()
	defer conn.Close()
	defer Recover()

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
	hub := newServerHub(newTunnel(conn))
	self.addHub(hub)
	defer self.removeHub(hub)

	err := hub.Start()
	if err != nil {
		Error("hub start failed:%s", err.Error())
		return
	}
	hub.Wait()
}

func (self *TunnelServer) listen() {
	defer self.wg.Done()

	for {
		conn, err := self.accept()
		if err != nil {
			Error("back server acceept failed:%s", err.Error())
			if opErr, ok := err.(*net.OpError); ok {
				if !opErr.Temporary() {
					break
				}
			}
			continue
		}
		Debug("back server, new connection from %v", conn.RemoteAddr())
		self.wg.Add(1)
		go self.handleClient(conn)
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
	self.rw.RLock()
	defer self.rw.RUnlock()

	for hub := range self.hubs {
		hub.Reload()
	}
	return nil
}

func (self *TunnelServer) Stop() {
	self.closeListener()
}

func (self *TunnelServer) Wait() {
	self.wg.Wait()
	Error("back hub quit")
}

func NewTunnelServer() *TunnelServer {
	tunnelServer := new(TunnelServer)
	tunnelServer.TcpServer.addr = options.Listen
	tunnelServer.hubs = make(map[*ServerHub]bool)
	return tunnelServer
}
