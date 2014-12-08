//
//   date  : 2014-07-16
//   author: xjdrew
//

package tunnel

import (
	"net"
	"sync"
)

type TunnelClient struct {
	ln  net.Listener
	hub *Hub
	wg  sync.WaitGroup
}

func (self *TunnelClient) createHub() error {
	conn, err := net.Dial("tcp", options.Server)
	if err != nil {
		return err
	}
	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
	self.hub = newHub(newTunnel(conn.(*net.TCPConn)))
	return err
}

func (self *TunnelClient) handleConn(conn *net.TCPConn) {
	defer self.wg.Done()
	defer conn.Close()
	defer Recover()

	linkid := self.hub.AcquireId()
	if linkid == 0 {
		Error("alloc linkid failed, source: %v", conn.RemoteAddr())
		return
	}
	defer self.hub.ReleaseId(linkid)

	Info("link(%d) create link, source: %v", linkid, conn.RemoteAddr())
	link := self.hub.NewLink(linkid)
	defer self.hub.ReleaseLink(linkid)

	link.SendCreate()
	link.Pump(conn)
}

func (self *TunnelClient) listen() {
	defer self.wg.Done()

	var err error
	self.ln, err = net.Listen("tcp", options.Listen)
	if err != nil {
		Panic("listen failed:%v", err)
	}

	for {
		conn, err := self.ln.Accept()
		if err != nil {
			Log("acceept failed:%s", err.Error())
			if opErr, ok := err.(*net.OpError); ok {
				if !opErr.Temporary() {
					break
				}
			}
			continue
		}
		Info("new connection from %v", conn.RemoteAddr())
		self.wg.Add(1)
		go self.handleConn(conn.(*net.TCPConn))
	}
}

func (self *TunnelClient) Start() error {
	err := self.createHub()
	if err != nil {
		return err
	}
	self.hub.Start()

	self.wg.Add(1)
	go self.listen()
	return nil
}

func (self *TunnelClient) Reload() error {
	return nil
}

func (self *TunnelClient) Stop() {
	self.ln.Close()
	self.hub.Close()
	Log("close tunnel client")
}

func (self *TunnelClient) Wait() {
	self.hub.Wait()
	self.wg.Wait()
	Log("tunnel client quit")
}

func (self *TunnelClient) Status() {
	self.hub.Status()
}

func NewTunnelClient() *TunnelClient {
	return &TunnelClient{}
}
