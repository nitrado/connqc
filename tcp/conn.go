package tcp

import (
	"fmt"
	"net"
)

var _ net.Conn = &Conn{}

// Conn contains the TCP connection.
type Conn struct {
	net.Conn
}

// NewConn returns a new TCP connection.
func NewConn(addr string) (*Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &Conn{
		conn,
	}, nil
}
