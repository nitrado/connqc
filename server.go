package signal

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
)

// Server handles connections from clients.
type Server struct {
	wg     sync.WaitGroup
	doneCh chan struct{}

	bufSize      int
	readTimeout  time.Duration
	writeTimeout time.Duration

	log *logger.Logger
}

// NewServer returns a server.
func NewServer(bufSize int, readTimeout, writeTimeout time.Duration, log *logger.Logger) *Server {
	return &Server{
		doneCh:       make(chan struct{}),
		bufSize:      bufSize,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		log:          log,
	}
}

// ServeTCP handles a TCP connection from a client.
// It writes back the request data as its response.
func (s *Server) ServeTCP(conn net.Conn) {
	s.wg.Add(1)
	defer func() {
		_ = conn.Close()
		s.wg.Done()
	}()

	go func() {
		<-s.doneCh
		_ = conn.Close()
	}()

	buf := make([]byte, s.bufSize)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(s.readTimeout))

		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			s.log.Error("Could not read request", lctx.Error("error", err))
			return
		}

		s.log.Debug("Received data", lctx.Str("data", string(buf[:n])))

		_ = conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))

		wn, err := conn.Write(buf[:n])
		if err != nil {
			s.log.Error("Could not write response", lctx.Error("error", err))
			return
		}
		if wn != n {
			s.log.Error("Unexpected write length", lctx.Int("expected", n), lctx.Int("actual", wn))
		}
	}
}

// Close closes a receiver once all connections are done.
func (s *Server) Close() error {
	close(s.doneCh)

	s.wg.Wait()

	return nil
}
