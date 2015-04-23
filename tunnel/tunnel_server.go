//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"io"
	"net"
	"sync"
)

type TunnelServer struct {
	ln   net.Listener
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

func (self *TunnelServer) handleConn(conn *net.TCPConn) {
	defer self.wg.Done()
	defer conn.Close()
	defer Recover()

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())

	// authenticate connection
	a := NewTaa(options.Secret)
	a.GenToken()

	challenge := a.GenCipherBlock(nil)
	Debug("challenge(%v), len %d, %v", conn.RemoteAddr(), len(challenge), challenge)
	if _, err := conn.Write(challenge); err != nil {
		Error("write challenge failed(%v):%s", conn.RemoteAddr(), err)
		return
	}

	token := make([]byte, TaaBlockSize)
	if _, err := io.ReadFull(conn, token); err != nil {
		Error("read token failed(%v):%s", conn.RemoteAddr(), err)
		return
	}

	Debug("token(%v), len %d, %v", conn.RemoteAddr(), len(token), token)
	if !a.VerifyCipherBlock(token) {
		Error("verify token failed(%v)", conn.RemoteAddr())
		return
	}

	hub := newServerHub(newTunnel(conn, a.GetRc4key()))
	self.addHub(hub)
	defer self.removeHub(hub)

	hub.Start()
}

func (self *TunnelServer) listen() {
	defer self.wg.Done()

	var err error
	self.ln, err = net.Listen("tcp", options.Listen)
	if err != nil {
		Panic("listen failed:%v", err)
	}

	for {
		conn, err := self.ln.Accept()
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
		tcpConn := conn.(*net.TCPConn)
		go self.handleConn(tcpConn)
	}
}

func (self *TunnelServer) Start() error {
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

func (self *TunnelServer) Wait() {
	self.wg.Wait()
	Error("back hub quit")
}

func (self *TunnelServer) Status() {
	self.rw.RLock()
	defer self.rw.RUnlock()

	for hub := range self.hubs {
		hub.Status()
	}
}

func NewTunnelServer() *TunnelServer {
	return &TunnelServer{
		hubs: make(map[*ServerHub]bool)}
}
