package udp

import (
	"fmt"
	"net"
)

// Connect returns a new UDP connection.
func Connect(addr string) (net.Conn, error) {
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return conn, err
}
