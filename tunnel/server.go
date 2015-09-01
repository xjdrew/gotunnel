//
//   date  : 2014-06-05
//   author: xjdrew
//

package tunnel

import (
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	laddr *net.TCPAddr
	baddr *net.TCPAddr

	secret string

	hubs map[*ServerHub]bool
	rw   sync.Mutex
}

func (server *Server) addHub(hub *ServerHub) {
	server.rw.Lock()
	server.hubs[hub] = true
	server.rw.Unlock()
}

func (server *Server) removeHub(hub *ServerHub) {
	server.rw.Lock()
	delete(server.hubs, hub)
	server.rw.Unlock()
}

func (server *Server) handleConn(conn *net.TCPConn) {
	defer conn.Close()
	defer Recover()

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())

	// authenticate connection
	a := NewTaa(server.secret)
	a.GenToken()

	challenge := a.GenCipherBlock(nil)
	Debug("challenge(%v), len %d, %v", conn.RemoteAddr(), len(challenge), challenge)
	if _, err := conn.Write(challenge); err != nil {
		Error("write challenge failed(%v):%s", conn.RemoteAddr(), err)
		return
	}

	token := make([]byte, TaaBlockSize)
	if _, err := io.ReadFull(conn, token); err != nil {
		Error("read token failed(%v):%s", conn.RemoteAddr(), err)
		return
	}

	Debug("token(%v), len %d, %v", conn.RemoteAddr(), len(token), token)
	if !a.VerifyCipherBlock(token) {
		Error("verify token failed(%v)", conn.RemoteAddr())
		return
	}

	tunnel := newTunnel(conn, a.GetRc4key())
	hub := newServerHub(tunnel, server.baddr)
	server.addHub(hub)
	defer server.removeHub(hub)

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
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Second * 180)
		go server.handleConn(conn)
	}
}

func (server *Server) Start() error {
	go server.listen()
	return nil
}

func (server *Server) Status() {
	for hub := range server.hubs {
		hub.Status()
	}
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
		hubs:   make(map[*ServerHub]bool),
	}
	return server, nil
}
