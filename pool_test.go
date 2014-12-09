package main

import "testing"
import "github.com/xjdrew/gotunnel/tunnel"

func TestPool(t *testing.T) {
	p := tunnel.NewMPool(64)
	a := p.Get()
	if p.Used() != 1 {
		t.Error("unexpected used:%d", p.Used)
	}
	p.Put(a)
	if p.Used() != 0 {
		t.Error("unexpected free:%d", p.Used)
	}
}
