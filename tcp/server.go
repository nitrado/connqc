package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

var testHookServerServe func(net.Listener)

// Handler handles TCP connections.
type Handler interface {
	Serve(conn net.PacketConn)
}

// Server serves TCP connections.
type Server struct {
	handler Handler
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
func (s *Server) Listen(ctx context.Context, addr string) error {
	lc := &net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listening: %w", err)
	}
	defer func() { _ = ln.Close() }()

	if testHookServerServe != nil {
		testHookServerServe(ln)
	}

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	var conn net.Conn
	for {
		conn, err = ln.Accept()

		var netErr net.Error
		switch {
		case err != nil && errors.Is(err, net.ErrClosed):
			return err
		case err != nil && errors.As(err, &netErr) && netErr.Timeout():
			return err
		case err != nil:
			return fmt.Errorf("accepting connection: %w", err)
		}

		go func() {
			defer func() { _ = conn.Close() }()

			s.handler.Serve(&packetConn{conn})
		}()
	}
}

var _ net.PacketConn = &packetConn{}

// packetConn makes a TCP connection act like an unbound connection
// to support the same interface that a UDP connection offers.
type packetConn struct {
	conn net.Conn
}

func (c *packetConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = c.conn.Read(p)
	return n, c.conn.RemoteAddr(), err
}

func (c *packetConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return c.conn.Write(p)
}

func (c *packetConn) Close() error {
	return c.conn.Close()
}

func (c *packetConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *packetConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *packetConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *packetConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
