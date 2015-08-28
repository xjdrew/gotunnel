//
//   date  : 2014-06-11
//   author: xjdrew
//
package tunnel

import (
	"net"
	"runtime"
)

const (
	MaxLinkPerTunnel = 1024
	PacketSize       = 8192
)

var (
	mpool = NewMPool(PacketSize)
)

type Service interface {
	Start() error
	Wait()
	Status()
}

type App struct {
	Listen  string
	Backend string // tunnel server or client
	Secret  string
	Tunnels uint // low level tunnel count; 0 if work as server

	laddr   *net.TCPAddr
	baddr   *net.TCPAddr
	service Service
}

func (app *App) Start() error {
	var err error
	if app.laddr, err = net.ResolveTCPAddr("tcp", app.Listen); err != nil {
		return err
	}

	if app.baddr, err = net.ResolveTCPAddr("tcp", app.Backend); err != nil {
		return err
	}

	if app.Tunnels == 0 {
		app.service = newServer(app)
	} else {
		app.service = newClient(app)
	}
	err = app.service.Start()
	return err
}

func (app *App) Wait() {
	app.service.Wait()
}

func (app *App) Status() {
	app.service.Status()
	LogStack("<status> num goroutine: %d", runtime.NumGoroutine())
}
