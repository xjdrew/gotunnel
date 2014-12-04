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

const SIG_STOP = syscall.Signal(34)
const SIG_RELOAD = syscall.Signal(35)
const SIG_STATUS = syscall.Signal(36)

func handleSignal(app *tunnel.App) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, SIG_STOP, SIG_RELOAD, SIG_STATUS, syscall.SIGTERM, syscall.SIGHUP)

	for sig := range c {
		switch sig {
		case SIG_STOP:
			app.Stop()
		case SIG_RELOAD:
			app.Reload()
		case SIG_STATUS:
			app.Status()
		default:
			tunnel.Log("catch signal:%v, ignore", sig)
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

	var rc4Key string
	flag.StringVar(&options.Listen, "listen", ":8001", "host:port gotunnel listen on")
	flag.StringVar(&options.Server, "server", "", "server address, empty if work as server")
	flag.IntVar(&options.LogLevel, "log", 1, "larger value for detail log")
	flag.StringVar(&rc4Key, "rc4", "the answer to life, the universe and everything", "rc4 key, disable if no key")
	//flag.IntVar(&options.Count, "count", 1, "underlayer tunnel count")
	flag.Usage = usage
	flag.Parse()

	options.Capacity = 20480
	options.RbufHw = 16
	options.RbufLw = 4
	options.PacketSize = 4096
	options.Rc4Key = []byte(rc4Key)

	if len(options.Rc4Key) > 256 {
		fmt.Println("rc4 key at most 256 bytes")
		os.Exit(1)
	}

	// will support multiple tunnel in future
	options.Count = 1
	if options.Count <= 0 || options.Count > 1024 {
		fmt.Println("tunnel count must be in range [1, 1024]")
		os.Exit(1)
	}

	if options.Server == "" {
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
	if options.Server == "" {
		app.Add(tunnel.NewTunnelServer())
	} else {
		app.Add(tunnel.NewTunnelClient())
	}

	err := app.Start()
	if err != nil {
		fmt.Printf("start gotunnel failed:%s\n", err.Error())
		return
	}
	go handleSignal(app)

	app.Wait()
}
