//
//   date  : 2015-12-28
//   author: xjdrew
//

package tunnel

import (
	"net"
)

func newListener(laddr string) (net.Listener, error) {
	return newTcpListener(laddr)
}

func dial(raddr string) (net.Conn, error) {
	return dialTcp(raddr)
}
