package main

import "testing"
import "github.com/xjdrew/gotunnel/tunnel"

func TestQos(t *testing.T) {
	highWater := 16
	lowWater := 4

	callEnterHighWatertimes := 0
	callEnterLowWatertimes := 0

	qos := tunnel.NewQos(highWater, lowWater, func() {
		callEnterHighWatertimes += 1
	}, func() {
		callEnterLowWatertimes += 1
	})

	qos.SetWater(highWater)
	qos.SetWater(lowWater)

	if callEnterHighWatertimes != 1 {
		t.Errorf("unexpected call enter low times:%d(%d)", callEnterHighWatertimes, 1)
	}
	if callEnterLowWatertimes != 1 {
		t.Errorf("unexpected call enter low times:%d(%d)", callEnterLowWatertimes, 1)
	}
	qos.SetWater(lowWater)
	qos.SetWater(highWater)
	qos.SetWater(highWater - 1)
	qos.SetWater(lowWater)
	if callEnterHighWatertimes != 2 {
		t.Errorf("unexpected call enter low times:%d(%d)", callEnterHighWatertimes, 2)
	}
	if callEnterLowWatertimes != 2 {
		t.Errorf("unexpected call enter low times:%d(%d)", callEnterLowWatertimes, 2)
	}
}

func TestQosRemote(t *testing.T) {
	highWater := 16
	lowWater := 4

	callEnterHighWatertimes := 0
	callEnterLowWatertimes := 0

	qos := tunnel.NewQos(highWater, lowWater, func() {
		callEnterHighWatertimes += 1
	}, func() {
		callEnterLowWatertimes += 1
	})

	qos.Balance()
	qos.SetRemoteFlag(true)
	go qos.SetRemoteFlag(false)
	qos.Balance()

	qos.SetRemoteFlag(true)
	go qos.Close()
	qos.Balance()

	if callEnterHighWatertimes != 0 {
		t.Errorf("unexpected call enter low times:%d(%d)", callEnterHighWatertimes, 0)
	}
	if callEnterLowWatertimes != 0 {
		t.Errorf("unexpected call enter low times:%d(%d)", callEnterLowWatertimes, 0)
	}
}
