package tcp

import (
	"fmt"
	"net"
)

// Connect returns a new TCP connection.
func Connect(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dialing: %w", err)
	}

	return conn, nil
}
