package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var remote string
	flag.StringVar(&remote, "remote", "127.0.0.1:7002", "remote server")
	flag.Usage = usage
	flag.Parse()

	conn, err := net.Dial("tcp", remote)
	if err != nil {
		log.Printf("connect server failed:%v", err)
		os.Exit(1)
	}
	w := conn.(*net.TCPConn)
	w.CloseRead()

	conn.Write([]byte(`hello
this is sender, a deaf, cann't hear any music
glad to talk to you,
best regard
`))
	w.CloseWrite()
}
