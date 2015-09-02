//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"net"
)

type Server struct {
	laddr *net.TCPAddr
	baddr *net.TCPAddr

	secret string
}

func (server *Server) handleConn(conn *net.TCPConn) {
	defer conn.Close()
	defer Recover()

	tunnel := newTunnel(conn)
	// authenticate connection
	a := NewTaa(server.secret)
	a.GenToken()

	challenge := a.GenCipherBlock(nil)
	if err := tunnel.Write(0, challenge); err != nil {
		Error("write challenge failed(%v):%s", tunnel, err)
		return
	}

	_, token, err := tunnel.Read()
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
	ln, err := net.ListenTCP("tcp", server.laddr)
	if err != nil {
		Panic("listen failed:%v", err)
	}

	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			Error("back server acceept failed:%s", err.Error())
			if opErr, ok := err.(*net.OpError); ok {
				if !opErr.Temporary() {
					break
				}
			}
			continue
		}
		Debug("back server, new connection from %v", conn.RemoteAddr())
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
	laddr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		return nil, err
	}

	baddr, err := net.ResolveTCPAddr("tcp", backend)
	if err != nil {
		return nil, err
	}

	server := &Server{
		laddr:  laddr,
		baddr:  baddr,
		secret: secret,
	}
	return server, nil
}
