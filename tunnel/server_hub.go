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
)

type Host struct {
	Addr   string
	Weight int
}

type Upstream struct {
	Hosts  []Host
	weight int
}

type ServerHub struct {
	*Hub
	upstream *Upstream
}

func (self *ServerHub) readSettings() (upstream *Upstream, err error) {
	fp, err := os.Open(options.ConfigFile)
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
		upstream.weight += host.Weight
	}

	Log("config:%v", upstream)
	return
}

func (self *ServerHub) chooseHost() (host *Host) {
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

func (self *ServerHub) handleLink(linkid uint16, link *Link) {
	defer self.Hub.wg.Done()
	defer Recover()

	host := self.chooseHost()
	if host == nil {
		Error("link(%d) choose host failed", linkid)
		self.SendLinkClose(linkid)
		return
	}

	dest, err := net.Dial("tcp", host.Addr)
	if err != nil {
		Error("link(%d) connect to host failed, host:%s, err:%v", linkid, host.Addr, err)
		self.SendLinkClose(linkid)
		return
	}

	Info("link(%d) new connection to %v", linkid, dest.RemoteAddr())
	link.Pump(dest.(*net.TCPConn))
}

func (self *ServerHub) Ctrl(cmd *CmdPayload) bool {
	linkid := cmd.Linkid
	switch cmd.Cmd {
	case LINK_CREATE:
		if self.setRWflag(linkid) {
			Info("link(%d) build link", linkid)
			self.Hub.wg.Add(1)
			link := self.NewLink(linkid)
			go self.handleLink(linkid, link)
		} else {
			Error("link(%d) id conflict", linkid)
			self.SendLinkClose(linkid)
		}
		return true
	}
	return false
}

func (self *ServerHub) Reload() error {
	Info("reload services")
	upstream, err := self.readSettings()
	if err != nil {
		Error("server hub load config file failed:%v", err)
		return err
	}
	self.upstream = upstream
	return nil
}

func (self *ServerHub) Start() error {
	err := self.Reload()
	if err != nil {
		return err
	}

	self.Hub.Start()
	return nil
}

func newServerHub(tunnel *Tunnel) *ServerHub {
	serverHub := new(ServerHub)
	hub := newHub(tunnel)
	hub.SetCtrlDelegate(serverHub)
	serverHub.Hub = hub
	return serverHub
}
