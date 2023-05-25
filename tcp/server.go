package tcp

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
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
	handler Handler

	inShutdown atomicBool

	mu        sync.Mutex
	listeners map[*net.Listener]struct{}
}

// NewServer returns a server with the given handler.
func NewServer(h Handler) (*Server, error) {
	if h == nil {
		return nil, errors.New("tcp: handler cannot be nil")
	}

	return &Server{
		handler:   h,
		listeners: map[*net.Listener]struct{}{},
	}, nil
}

// Listen listens to an address for new connections, passing them
// off to the handler in a goroutine.
func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	ln = &onceCloseListener{Listener: ln}
	defer func() { _ = ln.Close() }()

	if !s.addListener(&ln) {
		return ErrServerClosed
	}
	defer s.removeListener(&ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if s.inShutdown.isSet() {
				return ErrServerClosed
			}

			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				continue
			}

			return err
		}

		if s.inShutdown.isSet() {
			return ErrServerClosed
		}

		go s.handler.ServeTCP(conn)
	}
}

func (s *Server) addListener(ln *net.Listener) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.inShutdown.isSet() {
		return false
	}

	s.listeners[ln] = struct{}{}
	return true
}

func (s *Server) removeListener(ln *net.Listener) {
	s.mu.Lock()

	delete(s.listeners, ln)

	s.mu.Unlock()
}

// caller must hold s.mu.
func (s *Server) closeListeners() error {
	var err error
	for ln := range s.listeners {
		if cerr := (*ln).Close(); cerr != nil && err == nil {
			err = cerr
		}
		delete(s.listeners, ln)
	}
	return err
}

// Close closes the server.
func (s *Server) Close() error {
	s.inShutdown.set()

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.closeListeners()
}

type onceCloseListener struct {
	net.Listener
	once sync.Once
	err  error
}

func (l *onceCloseListener) close() {
	l.err = l.Listener.Close()
}

func (l *onceCloseListener) Close() error {
	l.once.Do(l.close)
	return l.err
}

type atomicBool int32

func (b *atomicBool) isSet() bool {
	return atomic.LoadInt32((*int32)(b)) != 0
}

func (b *atomicBool) set() {
	atomic.StoreInt32((*int32)(b), 1)
}
