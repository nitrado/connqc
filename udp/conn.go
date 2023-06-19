package udp

import (
	"fmt"
	"net"
)

var _ net.Conn = &Conn{}

// Conn contains the UDP connection.
type Conn struct {
	*net.UDPConn
}

// NewConn returns a new UDP connection.
func NewConn(addr string) (*Conn, error) {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &Conn{
		conn,
	}, nil
}
