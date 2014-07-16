//
//   date  : 2014-07-16
//   author: xjdrew
//

package tunnel

import (
	"math/rand"
	"net"
	"sync"
)

type FrontServer struct {
	TcpServer
	coors map[*Coor]bool
	wg    sync.WaitGroup
	rw    sync.RWMutex
}

func (self *FrontServer) addCoor(coor *Coor) {
	self.rw.Lock()
	self.coors[coor] = true
	self.rw.Unlock()
}

func (self *FrontServer) removeCoor(coor *Coor) {
	self.rw.Lock()
	delete(self.coors, coor)
	self.rw.Unlock()
}

func (self *FrontServer) chooseCoor() *Coor {
	self.rw.RLock()
	defer self.rw.RUnlock()

	n := len(self.coors)
	if n == 0 {
		return nil
	}

	v := rand.Intn(n)
	for coor := range self.coors {
		if v == 0 {
			return coor
		}
		v = v - 1
	}
	// impossible
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

	coor := self.chooseCoor()
	if coor == nil {
		Error("choose coor failed, source: %v", conn.RemoteAddr())
		conn.Close()
		return
	}

	linkid := coor.AcquireId()
	if linkid == 0 {
		Error("alloc linkid failed, source: %v", conn.RemoteAddr())
		conn.Close()
		return
	}

	ch := make(chan []byte, 256)
	err = coor.Set(linkid, ch)
	if err != nil {
		//impossible
		conn.Close()
		Error("set link failed, linkid:%d, source: %v", linkid, conn.RemoteAddr())
		return
	}

	Info("link(%d) create link, source: %v", linkid, conn.RemoteAddr())
	defer coor.ReleaseId(linkid)

	coor.SendLinkCreate(linkid)

	link := NewLink(linkid, conn)
	link.Pump(coor, ch)
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
	return nil
}

func (self *FrontServer) Stop() {
	self.closeListener()
}

func (self *FrontServer) Wait() {
	self.wg.Wait()
	Error("front door quit")
}

func NewFrontServer() *FrontServer {
	frontServer := new(FrontServer)
	frontServer.TcpServer.addr = options.FrontAddr
	frontServer.coors = make(map[*Coor]bool)
	return frontServer
}
