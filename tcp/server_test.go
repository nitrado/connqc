package tcp_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/hamba/testutils/retry"
	"github.com/nitrado/connqc/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer_ErrorsOnNilHandler(t *testing.T) {
	_, err := tcp.NewServer(nil)

	assert.Error(t, err)
}

func TestServer_Listen(t *testing.T) {
	_, conn := newTestServer(t, &echoHandler{})
	defer func() { _ = conn.Close() }()

	for i := 0; i < 3; i++ {
		msg := fmt.Sprintf("Hello %d", i)

		_, err := io.WriteString(conn, msg)
		require.NoError(t, err, "write error")

		got := make([]byte, 1024)
		n, err := conn.Read(got)
		require.NoError(t, err, "read error")

		assert.Equal(t, msg, string(got[:n]))
	}
}

func newTestServer(t testing.TB, h tcp.Handler) (*tcp.Server, net.Conn) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr := ln.Addr()
	_ = ln.Close()

	<-time.After(10 * time.Millisecond)

	srv, err := tcp.NewServer(h)
	require.NoError(t, err)

	go func() {
		err = srv.Listen(context.Background(), addr.String())
		if err != nil && err != net.ErrClosed {
			t.Fatal(err)
		}
	}()

	var conn net.Conn
	retry.Run(t, func(t *retry.SubT) {
		conn, err = net.Dial("tcp", addr.String())
		require.NoError(t, err)
	})

	return srv, conn
}

type echoHandler struct{}

func (e echoHandler) Serve(conn net.PacketConn) {
	buf := make([]byte, 512)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return
		}

		if _, err = conn.WriteTo(buf[:n], addr); err != nil {
			return
		}
	}
}
