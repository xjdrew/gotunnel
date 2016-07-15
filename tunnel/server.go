//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"net"
)

// server hub
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
		h.SendCmd(l.id, LINK_CLOSE)
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
			h.SendCmd(id, LINK_CLOSE)
		}
		return true
	case TUN_HEARTBEAT:
		h.SendCmd(id, TUN_HEARTBEAT)
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

// tunnel server
type Server struct {
	ln     net.Listener
	baddr  *net.TCPAddr
	secret string
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	defer Recover()

	tunnel := newTunnel(conn)
	// authenticate connection
	a := NewTaa(s.secret)
	a.GenToken()

	challenge := a.GenCipherBlock(nil)
	if err := tunnel.WritePacket(0, challenge); err != nil {
		Error("write challenge failed(%v):%s", tunnel, err)
		return
	}

	_, token, err := tunnel.ReadPacket()
	if err != nil {
		Error("read token failed(%v):%s", tunnel, err)
		return
	}

	if !a.VerifyCipherBlock(token) {
		Error("verify token failed(%v)", tunnel)
		return
	}

	tunnel.SetCipherKey(a.GetRc4key())
	h := newServerHub(tunnel, s.baddr)
	h.Start()
}

func (s *Server) Start() error {
	defer s.ln.Close()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				Log("acceept failed temporary: %s", netErr.Error())
				continue
			} else {
				return err
			}
		}
		Log("new connection from %v", conn.RemoteAddr())
		go s.handleConn(conn)
	}
}

func (s *Server) Status() {
}

// create a tunnel server
func NewServer(listen, backend, secret string) (*Server, error) {
	ln, err := newListener(listen)
	if err != nil {
		return nil, err
	}

	baddr, err := net.ResolveTCPAddr("tcp", backend)
	if err != nil {
		return nil, err
	}

	s := &Server{
		ln:     ln,
		baddr:  baddr,
		secret: secret,
	}
	return s, nil
}
