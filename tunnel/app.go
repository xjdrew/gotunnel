//
//   date  : 2014-06-11
//   author: xjdrew
//
package tunnel

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type Options struct {
	Listen     string
	Server     string // tunnel server or client
	Count      int    // tunnel count underlayer
	RbufHw     int    // recv buffer high water
	RbufLw     int    // recv buffer low water
	ConfigFile string
	LogLevel   int
	Rc4Key     []byte
	Capacity   uint16
	PacketSize uint16
}

var options *Options
var mpool *MPool

func init() {
	rand.Seed(time.Now().Unix())
}

type Service interface {
	Start() error
	Reload() error
	Stop()
	Wait()
}

type App struct {
	services []Service
	wg       sync.WaitGroup
}

func (self *App) Add(service Service) {
	self.services = append(self.services, service)
}

func (self *App) Start() error {
	mpool = NewMPool(int(options.PacketSize))
	for _, service := range self.services {
		err := service.Start()
		if err != nil {
			return err
		}
	}

	for _, service := range self.services {
		self.wg.Add(1)
		go func(s Service) {
			defer self.wg.Done()
			s.Wait()
			Info("service finish: %v", s)
		}(service)
	}
	return nil
}

func (self *App) Reload() {
	for _, service := range self.services {
		service.Reload()
	}
}

func (self *App) Stop() {
	for _, service := range self.services {
		service.Stop()
	}
}

func (self *App) Wait() {
	self.wg.Wait()
}

func (self *App) Status() {
	Log("num goroutine: %d, pool %d(%d)", runtime.NumGoroutine(), mpool.Used(), mpool.Alloced())
}

func NewApp(o *Options) *App {
	options = o
	return new(App)
}
