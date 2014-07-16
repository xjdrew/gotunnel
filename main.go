//
//   date  : 2014-06-04
//   author: xjdrew
//

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xjdrew/gotunnel/tunnel"
)

const SIG_STOP = syscall.Signal(34)
const SIG_RELOAD = syscall.Signal(35)
const SIG_STATUS = syscall.Signal(36)

func handleSignal(app *tunnel.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, SIG_STOP, SIG_RELOAD, SIG_STATUS, syscall.SIGTERM)

	for sig := range c {
		switch sig {
		case SIG_STOP:
			app.Stop()
		case SIG_RELOAD:
			app.Reload()
		case SIG_STATUS:
			fmt.Println("catch sigstatus, ignore")
		case syscall.SIGTERM:
			fmt.Println("catch sigterm, ignore")
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [configFile]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func argsCheck() *tunnel.Options {
	var options tunnel.Options

	var tgw string
	var rc4Key string
	flag.BoolVar(&options.TunnelServer, "tunnel_server", false, "work as tunnel server or client")
	flag.BoolVar(&options.Front, "front", false, "work as front door or back door")
	flag.StringVar(&tgw, "tgw", "", "tgw header")
	flag.StringVar(&rc4Key, "rc4", "the answer to life, the universe and everything", "rc4 key, disable if no key")
	flag.StringVar(&options.FrontAddr, "front_addr", "0.0.0.0:8001", "front door address(0.0.0.0:8001)")
	flag.StringVar(&options.TunnelAddr, "tunnel_addr", "0.0.0.0:8002", "tunnel door address(0.0.0.0:8002)")
	flag.IntVar(&options.LogLevel, "log", 1, "larger value for detail log")
	flag.Usage = usage
	flag.Parse()

	options.Capacity = 65535
	options.Tgw = bytes.ToLower([]byte(tgw))
	options.Rc4Key = []byte(rc4Key)

	if len(options.Rc4Key) > 256 {
		fmt.Println("rc4 key at most 256 bytes")
		os.Exit(1)
	}

	if !options.Front {
		args := flag.Args()
		if len(args) < 1 {
			usage()
		} else {
			options.ConfigFile = args[0]
		}
	}
	return &options
}

func main() {
	// parse argument and check
	options := argsCheck()

	app := tunnel.NewApp(options)
	if options.TunnelServer {
		var tunnelServer *tunnel.TunnelServer
		if options.Front {
			tunnelServer = tunnel.NewTunnelServer(tunnel.NewFrontDoor)
		} else {
			tunnelServer = tunnel.NewTunnelServer(tunnel.NewBackDoor)
		}
		app.Add(tunnelServer)
	} else {
		var tunnelClient *tunnel.TunnelClient
		if options.Front {
			tunnelClient = tunnel.NewTunnelClient(tunnel.NewFrontDoor)
		} else {
			tunnelClient = tunnel.NewTunnelClient(tunnel.NewBackDoor)
		}
		app.Add(tunnelClient)
	}

	err := app.Start()
	if err != nil {
		fmt.Printf("start gotunnel failed:%s\n", err.Error())
		return
	}
	go handleSignal(app)

	app.Wait()
}
