package tcp_test

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"dev.marbis.net/marbis/signal/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer_ErrorsOnNilHandler(t *testing.T) {
	_, err := tcp.NewServer(nil)

	assert.Error(t, err)
}

func TestServer_Listen(t *testing.T) {
	addr, srv := newTestServer(t, echoHandler{})
	t.Cleanup(func() { _ = srv.Close() })

	conn, err := net.Dial("tcp", addr.String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	for i := 0; i < 3; i++ {
		msg := fmt.Sprintf("Hello %d", i)

		_, err = io.WriteString(conn, msg)
		require.NoError(t, err, "write error")

		got := make([]byte, 1024)
		n, err := conn.Read(got)
		require.NoError(t, err, "read error")

		assert.Equal(t, msg, string(got[:n]))
	}
}

func newTestServer(t testing.TB, h tcp.Handler) (net.Addr, *tcp.Server) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := ln.Addr()
	_ = ln.Close()

	srv, err := tcp.NewServer(h)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		err = srv.Listen(addr.String())
		if err != nil && err != tcp.ErrServerClosed {
			t.Fatal(err)
		}
	}()

	wg.Wait()

	return addr, srv
}

type echoHandler struct{}

func (e echoHandler) ServeTCP(conn net.Conn) {
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		if _, err = conn.Write(buf[:n]); err != nil {
			return
		}
	}
}
