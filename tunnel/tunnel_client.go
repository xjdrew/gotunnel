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
	ln   net.Listener
	hubs []*Hub
	off  int // current hub
	wg   sync.WaitGroup
}

func (self *TunnelClient) createHub() (hub *Hub, err error) {
	conn, err := net.Dial("tcp", options.Server)
	if err != nil {
		return
	}
	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())
	hub = newHub(newTunnel(conn.(*net.TCPConn)))
	return
}

func (self *TunnelClient) handleConn(hub *Hub, conn *net.TCPConn) {
	defer conn.Close()
	defer Recover()

	linkid := hub.AcquireId()
	if linkid == 0 {
		Error("alloc linkid failed, source: %v", conn.RemoteAddr())
		return
	}
	defer hub.ReleaseId(linkid)

	Info("link(%d) create link, source: %v", linkid, conn.RemoteAddr())
	link := hub.NewLink(linkid)
	defer hub.ReleaseLink(linkid)

	link.SendCreate()
	link.Pump(conn)
}

func (self *TunnelClient) fetchHub() *Hub {
	for i := 0; i < len(self.hubs); i++ {
		hub := self.hubs[self.off]
		self.off = (self.off + 1) % len(self.hubs)
		if hub != nil {
			return hub
		}
	}
	Panic("no active tunnel")
	return nil
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
		hub := self.fetchHub()
		go self.handleConn(hub, conn.(*net.TCPConn))
	}
}

func (self *TunnelClient) Start() error {
	for i := 0; i < len(self.hubs); i++ {
		hub, err := self.createHub()
		if err != nil {
			return err
		}
		self.hubs[i] = hub

		go func(index int) {
			Recover()

			hub.Start()
			for {
				Error("tunnel %d disconnected", index)
				self.hubs[index] = nil
				hub, err := self.createHub()
				if err != nil {
					Error("tunnel %d reconnect failed", index)
					continue
				}
				Error("tunnel %d reconnect succeed", index)
				self.hubs[index] = hub
				hub.Start()
			}
		}(i)
	}

	self.wg.Add(1)
	go self.listen()
	return nil
}

func (self *TunnelClient) Reload() error {
	return nil
}

func (self *TunnelClient) Wait() {
	self.wg.Wait()
	Log("tunnel client quit")
}

func (self *TunnelClient) Status() {
	for _, hub := range self.hubs {
		hub.Status()
	}
}

func NewTunnelClient() *TunnelClient {
	count := 1
	if options.TunnelCount > 0 {
		count = options.TunnelCount
	}
	return &TunnelClient{
		hubs: make([]*Hub, count)}
}
