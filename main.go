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
	"syscall"

	"github.com/xjdrew/gotunnel/tunnel"
)

const SIG_STATUS = syscall.Signal(36)

func handleSignal(app *tunnel.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, SIG_STATUS, syscall.SIGTERM, syscall.SIGHUP)

	for sig := range c {
		switch sig {
		case SIG_STATUS:
			app.Status()
		default:
			tunnel.Log("catch signal:%v, ignore", sig)
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
	tunnels := flag.Uint("tunnels", 1, "low level tunnel count, 0 if work as server")
	flag.Int64Var(&tunnel.Timeout, "timeout", 1, "tunnel read/write timeout")
	flag.UintVar(&tunnel.LogLevel, "log", 1, "log level")

	flag.Usage = usage
	flag.Parse()

	app := &tunnel.App{
		Listen:  *laddr,
		Backend: *baddr,
		Secret:  *secret,
		Tunnels: *tunnels,
	}
	err := app.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "start failed:%s\n", err.Error())
		return
	}
	go handleSignal(app)

	app.Wait()
}
