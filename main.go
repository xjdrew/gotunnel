//
//   date  : 2014-06-04
//   author: xjdrew
//

package main

import (
	"bytes"
	"flag"
	"fmt"

	// "log/syslog"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

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

const SIG_STOP = syscall.Signal(34)
const SIG_RELOAD = syscall.Signal(35)
const SIG_STATUS = syscall.Signal(36)

func handleSignal(app *App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, SIG_STOP, SIG_RELOAD, SIG_STATUS, syscall.SIGTERM)

	for sig := range c {
		Info("catch signial:%d", sig)
		switch sig {
		case SIG_STOP:
			app.Stop()
		case SIG_RELOAD:
			app.Reload()
		case SIG_STATUS:
			Info("catch sigstatus, ignore")
		case syscall.SIGTERM:
			Info("catch sigterm, ignore")
		}
	}
}

type Options struct {
	gate       bool
	capacity   uint16
	frontAddr  string
	backAddr   string
	configFile string
	logLevel   int
	tgw        []byte
	rc4Key     []byte
}

var options Options

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [configFile]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func argsCheck() {
	var tgw string
	var rc4Key string
	flag.BoolVar(&options.gate, "gate", false, "as gate or node")
	flag.StringVar(&tgw, "tgw", "", "tgw header")
	flag.StringVar(&rc4Key, "rc4", "the answer to life, the universe and everything", "rc4 key, disable if no key")
	flag.StringVar(&options.frontAddr, "front_addr", "0.0.0.0:8001", "front door address(0.0.0.0:8001)")
	flag.StringVar(&options.backAddr, "back_addr", "0.0.0.0:8002", "back door address(0.0.0.0:8002)")
	flag.IntVar(&options.logLevel, "log", 1, "larger value for detail log")
	flag.Usage = usage
	flag.Parse()

	options.capacity = 65535
	options.tgw = bytes.ToLower([]byte(tgw))
	options.rc4Key = []byte(rc4Key)

	if len(options.rc4Key) > 256 {
		Error("rc4 key at most 256 bytes")
		os.Exit(1)
	}

	if !options.gate {
		args := flag.Args()
		if len(args) < 1 {
			usage()
		} else {
			options.configFile = args[0]
		}
	}
}

func main() {
	// parse argument and check
	argsCheck()

	app := new(App)
	if options.gate {
		backDoor := NewBackServer()
		app.Add(backDoor)
	} else {
		backClient := NewBackClient()
		app.Add(backClient)
	}

	err := app.Start()
	if err != nil {
		Error("start gotunnel failed:%s", err.Error())
		return
	}
	go handleSignal(app)

	app.Wait()
}
