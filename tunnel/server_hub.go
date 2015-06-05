//
//   date  : 2014-06-06
//   author: xjdrew
//

package tunnel

import (
	"net"
	"time"
)

type ServerHub struct {
	*Hub
	app *App
}

func (self *ServerHub) handleLink(linkid uint16, link *Link) {
	defer self.Hub.ReleaseLink(linkid)
	defer Recover()

	conn, err := net.DialTCP("tcp", nil, self.app.baddr)
	if err != nil {
		Error("link(%d) connect to backend failed, err:%v", linkid, err)
		link.SendClose()
		return
	}

	Info("link(%d) new connection to %v", linkid, conn.RemoteAddr())

	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 60)
	link.Pump(conn)
}

func (self *ServerHub) Ctrl(cmd *Cmd) bool {
	linkid := cmd.Linkid
	switch cmd.Cmd {
	case LINK_CREATE:
		link := self.NewLink(linkid)
		if link != nil {
			Info("link(%d) build link", linkid)
			go self.handleLink(linkid, link)
		} else {
			Error("link(%d) id conflict", linkid)
			self.Send(LINK_CLOSE, linkid, nil)
		}
		return true
	}
	return false
}

func newServerHub(tunnel *Tunnel, app *App) *ServerHub {
	ServerHub := new(ServerHub)
	ServerHub.app = app
	hub := newHub(tunnel, false)
	hub.SetCtrlDelegate(ServerHub)
	ServerHub.Hub = hub
	return ServerHub
}
