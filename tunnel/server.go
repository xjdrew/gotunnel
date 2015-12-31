//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"net"
)

type Server struct {
	ln     net.Listener
	baddr  *net.TCPAddr
	secret string
}

func (server *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	defer Recover()

	tunnel := newTunnel(conn)
	// authenticate connection
	a := NewTaa(server.secret)
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
	hub := newServerHub(tunnel, server.baddr)
	hub.Start()
}

func (server *Server) listen() {
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			Error("acceept failed:%s", err.Error())
			break
		}
		Log("new connection from %v", conn.RemoteAddr())
		go server.handleConn(conn)
	}
}

func (server *Server) Start() error {
	go server.listen()
	return nil
}

func (server *Server) Status() {
}

func NewServer(listen, backend, secret string) (*Server, error) {
	ln, err := newListener(listen)
	if err != nil {
		return nil, err
	}

	baddr, err := net.ResolveTCPAddr("tcp", backend)
	if err != nil {
		return nil, err
	}

	server := &Server{
		ln:     ln,
		baddr:  baddr,
		secret: secret,
	}
	return server, nil
}
