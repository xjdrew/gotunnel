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
	app  *App
	hubs map[*ServerHub]bool
	rw   sync.Mutex
	wg   sync.WaitGroup
}

func (self *Server) addHub(hub *ServerHub) {
	self.rw.Lock()
	self.hubs[hub] = true
	self.rw.Unlock()
}

func (self *Server) removeHub(hub *ServerHub) {
	self.rw.Lock()
	delete(self.hubs, hub)
	self.rw.Unlock()
}

func (self *Server) handleConn(conn *net.TCPConn) {
	defer self.wg.Done()
	defer conn.Close()
	defer Recover()

	Info("create tunnel: %v <-> %v", conn.LocalAddr(), conn.RemoteAddr())

	// authenticate connection
	a := NewTaa(self.app.Secret)
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

	hub := newServerHub(newTunnel(conn, a.GetRc4key()), self.app)
	self.addHub(hub)
	defer self.removeHub(hub)

	hub.Start()
}

func (self *Server) listen() {
	defer self.wg.Done()

	ln, err := net.ListenTCP("tcp", self.app.laddr)
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
		self.wg.Add(1)
		go self.handleConn(conn)
	}
}

func (self *Server) Start() error {
	self.wg.Add(1)
	go self.listen()
	return nil
}

func (self *Server) Wait() {
	self.wg.Wait()
	Error("back hub quit")
}

func (self *Server) Status() {
	for hub := range self.hubs {
		hub.Status()
	}
}

func newServer(app *App) *Server {
	return &Server{
		app:  app,
		hubs: make(map[*ServerHub]bool),
	}
}
