package tcp

import (
	"errors"
	"fmt"
	"net"
)

// ErrServerClosed is returned when a connection is
// attempted on a closed server.
var ErrServerClosed = errors.New("tcp: server closed")

// Handler handles TCP connections.
type Handler interface {
	ServeTCP(conn net.Conn)
}

// Server serves TCP connections.
type Server struct {
	handler  Handler
	listener net.Listener
}

// NewServer returns a server with the given handler.
func NewServer(h Handler) (*Server, error) {
	if h == nil {
		return nil, errors.New("tcp: handler cannot be nil")
	}

	return &Server{
		handler: h,
	}, nil
}

// Listen listens to an address for new connections, passing them
// off to the handler in a goroutine.
func (s *Server) Listen(addr string) error {
	if s.listener != nil {
		return fmt.Errorf("already listening")
	}

	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer func() { _ = s.listener.Close() }()

	for {
		conn, err := s.listener.Accept()

		var netErr net.Error
		switch {
		case err != nil && errors.Is(err, net.ErrClosed):
			return ErrServerClosed
		case err != nil && errors.As(err, &netErr) && netErr.Timeout():
			return err
		case err != nil:
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		go func() {
			defer func() { _ = conn.Close() }()

			s.handler.ServeTCP(conn)
		}()
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}
