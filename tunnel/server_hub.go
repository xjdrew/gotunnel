//
//   date  : 2014-06-06
//   author: xjdrew
//

package tunnel

import (
	"net"
)

type ServerHub struct {
	*Hub
	baddr *net.TCPAddr
}

func (hub *ServerHub) handleLink(link *Link) {
	defer Recover()

	conn, err := net.DialTCP("tcp", nil, hub.baddr)
	if err != nil {
		Error("link(%d) connect to backend failed, err:%v", link.id, err)
		link.SendClose()
		return
	}

	Info("link(%d) new connection to %v", link.id, conn.RemoteAddr())
	link.Pump(conn)
}

func (hub *ServerHub) onCtrl(cmd Cmd) bool {
	linkid := cmd.Linkid
	switch cmd.Cmd {
	case LINK_CREATE:
		link := newLink(linkid, hub.Hub)
		if link != nil {
			Info("link(%d) new link", linkid)
			go hub.handleLink(link)
		} else {
			Error("link(%d) id conflict")
			hub.Send(LINK_CLOSE, linkid, nil)
		}
		return true
	}
	return false
}

func newServerHub(tunnel *Tunnel, baddr *net.TCPAddr) *ServerHub {
	hub := &ServerHub{
		Hub:   newHub(tunnel),
		baddr: baddr,
	}
	hub.Hub.onCtrlFilter = hub.onCtrl
	return hub
}
