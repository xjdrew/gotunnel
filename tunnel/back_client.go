//
//   date  : 2014-06-06
//   author: xjdrew
//

package tunnel

import (
	"encoding/json"
	"math/rand"
	"net"
	"os"
	"sync"
)

type Host struct {
	Addr   string
	Weight int

	addr *net.TCPAddr
}

type Upstream struct {
	Hosts  []Host
	weight int
}

type BackClient struct {
	configFile string
	backAddr   string
	wg         sync.WaitGroup
	upstream   *Upstream
	coor       *Coor
}

func (self *BackClient) readSettings() (upstream *Upstream, err error) {
	fp, err := os.Open(self.configFile)
	if err != nil {
		Error("open config file failed:%s", err.Error())
		return
	}
	defer fp.Close()

	upstream = new(Upstream)
	dec := json.NewDecoder(fp)
	err = dec.Decode(upstream)
	if err != nil {
		Error("decode config file failed:%s", err.Error())
		return
	}

	for i := range upstream.Hosts {
		host := &upstream.Hosts[i]
		host.addr, err = net.ResolveTCPAddr("tcp", host.Addr)
		if err != nil {
			Error("resolve local addr failed:%s", err.Error())
			return
		}
		upstream.weight += host.Weight
	}

	Info("config:%v", upstream)
	return
}

func (self *BackClient) chooseHost() (host *Host) {
	upstream := self.upstream
	if upstream.weight <= 0 {
		return
	}
	v := rand.Intn(upstream.weight)
	for _, h := range upstream.Hosts {
		if h.Weight >= v {
			host = &h
			break
		}
		v -= h.Weight
	}
	return
}

func (self *BackClient) handleLink(linkid uint16, ch chan []byte) {
	defer self.wg.Done()

	host := self.chooseHost()
	if host == nil {
		Error("link(%d) choose host failed", linkid)
		self.coor.Reset(linkid)
		self.coor.SendLinkDestory(linkid)
		return
	}

	dest, err := net.DialTCP("tcp", nil, host.addr)
	if err != nil {
		Error("link(%d) connect to host failed, host:%s, err:%v", linkid, host.Addr, err)
		self.coor.Reset(linkid)
		self.coor.SendLinkDestory(linkid)
		return
	}

	Info("link(%d) new connection to %v", linkid, dest.RemoteAddr())
	link := NewLink(linkid, dest)
	link.Pump(self.coor, ch)
}

func (self *BackClient) ctrl(cmd *CmdPayload) bool {
	linkid := cmd.Linkid
	switch cmd.Cmd {
	case LINK_CREATE:
		ch := make(chan []byte, 256)
		err := self.coor.Set(linkid, ch)
		if err != nil {
			Error("build link failed, linkid:%d, error:%s", linkid, err)
			self.coor.SendLinkDestory(linkid)
			return true
		}
		Info("link(%d) build link", linkid)
		self.wg.Add(1)
		go self.handleLink(linkid, ch)
		return true
	default:
		return false
	}
}

func (self *BackClient) pump() {
	defer self.wg.Done()
	self.coor.Start()
	self.coor.Wait()
}

func (self *BackClient) Start() error {
	upstream, err := self.readSettings()
	if err != nil {
		return err
	}
	self.upstream = upstream

	addr, err := net.ResolveTCPAddr("tcp", self.backAddr)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}

	err = writeTGW(conn)
	if err != nil {
		Error("write tgw failed")
		return err
	}

	self.coor = NewCoor(NewTunnel(conn), self)

	self.wg.Add(1)
	go self.pump()
	return nil
}

func (self *BackClient) Reload() error {
	Info("reload services")
	upstream, err := self.readSettings()
	if err != nil {
		Error("back client reload failed:%v", err)
		return err
	}
	self.upstream = upstream
	return nil
}

func (self *BackClient) Stop() {
}

func (self *BackClient) Wait() {
	self.wg.Wait()
}

func NewBackClient() *BackClient {
	backClient := new(BackClient)
	backClient.configFile = options.ConfigFile
	backClient.backAddr = options.BackAddr
	return backClient
}
