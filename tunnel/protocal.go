//
//   date  : 2015-08-27
//   author: xjdrew
//

package tunnel

/*
client send authenticator to server:
type Authenticator struct {
	Nonce    uint64 // radmon value
	Checksum uint64 // checksum of Nonce and secret
}

If Checksum can be verified, server will response with a ticket as below:
type Ticket struct {
}

// tunnel wire protocal
type tunnelHeader struct {
	Linkid uint16
}

// authToken uses for authenticate client
type authToken struct {
}

// sessionToken uses for creating new tunnel, it's generated
// once for one tunnel
type sessionToken struct {
	Id    uint64 // client identifier
	Nonce uint64 // replay protection, must be strictly monotonically increase
}
*/
