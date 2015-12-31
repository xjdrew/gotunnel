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

func (h *ServerHub) handleLink(l *link) {
	defer Recover()
	defer h.deleteLink(l.id)

	conn, err := net.DialTCP("tcp", nil, h.baddr)
	if err != nil {
		Error("link(%d) connect to backend failed, err:%v", l.id, err)
		h.Send(LINK_CLOSE, l.id, nil)
		h.deleteLink(l.id)
		return
	}

	h.startLink(l, conn)
}

func (h *ServerHub) onCtrl(cmd Cmd) bool {
	id := cmd.Id
	switch cmd.Cmd {
	case LINK_CREATE:
		l := h.createLink(id)
		if l != nil {
			go h.handleLink(l)
		} else {
			h.Send(LINK_CLOSE, id, nil)
		}
		return true
	}
	return false
}

func newServerHub(tunnel *Tunnel, baddr *net.TCPAddr) *ServerHub {
	h := &ServerHub{
		Hub:   newHub(tunnel),
		baddr: baddr,
	}
	h.Hub.onCtrlFilter = h.onCtrl
	return h
}
