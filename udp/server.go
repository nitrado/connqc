// Package udp implements a UDP server.
package udp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

// Handler handles UDP connections.
type Handler interface {
	Serve(conn net.PacketConn)
}

// Server serves UDP connections.
type Server struct {
	handler Handler
}

// NewServer returns a server with the given handler.
func NewServer(h Handler) (*Server, error) {
	if h == nil {
		return nil, errors.New("udp: handler cannot be nil")
	}

	return &Server{
		handler: h,
	}, nil
}

// Listen listens to an address for new connections, passing them
// off to the handler in a goroutine.
func (s *Server) Listen(ctx context.Context, addr string) error {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	ln, err := net.ListenUDP(laddr.Network(), laddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer func() { _ = ln.Close() }()

	go s.handler.Serve(&gracefulRead{conn: ln})

	<-ctx.Done()

	return nil
}

var _ net.PacketConn = &gracefulRead{}

// gracefulRead represents a net.PacketConn where you can read and write to an address.
//
// To avoid a recurring read timeout when the UDP connection is unused, we store the activity (active = true)
// whenever we successfully read. This way we can escalate read errors only once and after we have seen activity.
type gracefulRead struct {
	active       bool
	readDeadline time.Duration
	conn         *net.UDPConn
}

func (g *gracefulRead) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	var netErr net.Error
	for {
		n, addr, err = g.conn.ReadFrom(p)

		switch {
		case err != nil && errors.As(err, &netErr) && netErr.Timeout() && !g.active:
			_ = g.conn.SetReadDeadline(time.Now().Add(g.readDeadline))
			continue
		case err != nil && errors.As(err, &netErr) && netErr.Timeout() && g.active:
			g.active = false
		case err == nil:
			g.active = true
		}

		return n, addr, err
	}
}

func (g *gracefulRead) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return g.conn.WriteTo(p, addr)
}

func (g *gracefulRead) Close() error {
	return g.conn.Close()
}

func (g *gracefulRead) LocalAddr() net.Addr {
	return g.conn.LocalAddr()
}

func (g *gracefulRead) SetDeadline(t time.Time) error {
	return g.conn.SetDeadline(t)
}

func (g *gracefulRead) SetReadDeadline(t time.Time) error {
	g.readDeadline = time.Until(t)
	return g.conn.SetReadDeadline(t)
}

func (g *gracefulRead) SetWriteDeadline(t time.Time) error {
	return g.conn.SetWriteDeadline(t)
}
