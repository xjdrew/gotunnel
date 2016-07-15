//
//   date  : 2014-06-04
//   author: xjdrew
//

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/xjdrew/gotunnel/tunnel"
)

type Service interface {
	Start() error
	Status()
}

func handleSignal(app Service) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	for sig := range c {
		switch sig {
		case syscall.SIGHUP:
			app.Status()
			tunnel.Log("total goroutines:%d", runtime.NumGoroutine())
		default:
			tunnel.Log("catch signal:%v, exit", sig)
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	laddr := flag.String("listen", ":8001", "listen address")
	baddr := flag.String("backend", "127.0.0.1:1234", "backend address")
	secret := flag.String("secret", "the answer to life, the universe and everything", "tunnel secret")
	tunnels := flag.Uint("tunnels", 0, "low level tunnel count, 0 if work as server")
	flag.IntVar(&tunnel.Timeout, "timeout", 30, "tunnel read/write timeout")
	flag.UintVar(&tunnel.LogLevel, "log", 1, "log level")
	flag.BoolVar(&tunnel.Udt, "udt", false, "udt tunnel")

	flag.Usage = usage
	flag.Parse()

	var app Service
	var err error
	if *tunnels == 0 {
		app, err = tunnel.NewServer(*laddr, *baddr, *secret)
	} else {
		app, err = tunnel.NewClient(*laddr, *baddr, *secret, *tunnels)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "create service failed:%s\n", err.Error())
		return
	}

	// waiting for signal
	go handleSignal(app)

	// start app
	tunnel.Log("exit: %v", app.Start())
}
