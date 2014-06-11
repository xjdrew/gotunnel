//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

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
	self.closeListener()
}

func (self *FrontServer) listen() {
	defer self.wg.Done()
	for {
		conn, err := self.accept()
		if err != nil {
			Error("front server acceept failed:%s", err.Error())
			break
		}
		Debug("front server, new connection from %v", conn.RemoteAddr())
		self.wg.Add(1)
		go self.handleClient(conn)
	}
}

func (self *FrontServer) Start() error {
	err := self.buildListener()
	if err != nil {
		return err
	}

	self.wg.Add(1)
	go self.listen()
	self.wg.Add(1)
	go self.pump()

	return nil
}

func (self *FrontServer) handleClient(conn *net.TCPConn) {
	defer self.wg.Done()

	// try skip tgw
	err := skipTGW(conn)
	if err != nil {
		Error("skip tgw failed, source: %v", conn.RemoteAddr())
		conn.Close()
		return
	}

	linkid := self.coor.AcquireId()
	if linkid == 0 {
		Error("alloc linkid failed, source: %v", conn.RemoteAddr())
		conn.Close()
		return
	}

	ch := make(chan []byte, 256)
	err = self.coor.Set(linkid, ch)
	if err != nil {
		//impossible
		conn.Close()
		Error("set link failed, linkid:%d, source: %v", linkid, conn.RemoteAddr())
		return
	}

	Info("link(%d) create link, source: %v", linkid, conn.RemoteAddr())
	defer self.coor.ReleaseId(linkid)

	self.coor.SendLinkCreate(linkid)

	link := NewLink(linkid, conn)
	link.Pump(self.coor, ch)
}

func (self *FrontServer) Stop() {
	self.closeListener()
}

func (self *FrontServer) Wait() {
	self.wg.Wait()
	Error("front door quit")
}

func NewFrontServer(tunnel *Tunnel) *FrontServer {
	frontServer := new(FrontServer)
	frontServer.coor = NewCoor(tunnel, nil)
	frontServer.TcpServer.addr = options.FrontAddr
	return frontServer
}
