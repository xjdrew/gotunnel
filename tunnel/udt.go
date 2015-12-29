//
//   date  : 2015-12-28
//   author: xjdrew
//

package tunnel

import (
	"net"

	udt "github.com/xjdrew/go-udtwrapper/udt"
)

// create a udt listener for server
func newUdtListener(laddr string) (net.Listener, error) {
	ln, err := udt.Listen("udt", laddr)
	if err != nil {
		return nil, err
	}
	return ln, err
}

// for client
func dialUdt(raddr string) (net.Conn, error) {
	return udt.Dial("udt", raddr)
}
