//
//   date  : 2015-12-28
//   author: xjdrew
//

package tunnel

import (
	"net"
)

func newListener(laddr string) (net.Listener, error) {
	if Udt {
		return newUdtListener(laddr)
	} else {
		return newTcpListener(laddr)
	}
}

func dial(raddr string) (net.Conn, error) {
	if Udt {
		return dialUdt(raddr)
	} else {
		return dialTcp(raddr)
	}
}
