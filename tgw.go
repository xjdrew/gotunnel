package main

import (
	"bytes"
	"errors"
	"net"
)

var ErrWrongTGW = errors.New("wrong tgw header")

func skipTGW(conn *net.TCPConn) (err error) {
	sz := len(options.tgw)
	if sz == 0 {
		return
	}

	buf := make([]byte, sz)
	s := 0
	for s < sz {
		var n int
		n, err = conn.Read(buf[s:])
		if err != nil {
			return
		}
		s += n
	}
	if bytes.Compare(bytes.ToLower(buf), options.tgw) == 0 {
		return
	}
	return ErrWrongTGW
}

func writeTGW(conn *net.TCPConn) (err error) {
	sz := len(options.tgw)
	if sz == 0 {
		return
	}
	s := 0
	for s < sz {
		var n int
		n, err = conn.Write(options.tgw[s:])
		if err != nil {
			return
		}
		s += n
	}
	return
}
