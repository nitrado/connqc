package udp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/hamba/testutils/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer_ErrorsOnNilHandler(t *testing.T) {
	_, err := NewServer(nil)

	assert.Error(t, err)
}

func TestServer_Listen(t *testing.T) {
	_, conn := newTestServer(t, &echoHandler{})
	t.Cleanup(func() { _ = conn.Close() })

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

func newTestServer(t testing.TB, h Handler) (*Server, net.Conn) {
	t.Helper()

	lnCh := make(chan *net.UDPConn, 1)
	setTestHookServerServe(func(ln *net.UDPConn) {
		lnCh <- ln
	})
	t.Cleanup(func() { setTestHookServerServe(nil) })

	srv, err := NewServer(h)
	require.NoError(t, err)

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		err = srv.Listen(ctx, "localhost:0")
		if err != nil && !errors.Is(err, net.ErrClosed) {
			t.Fatal(err)
		}
	}()

	ln := <-lnCh

	laddr, err := net.ResolveUDPAddr(ln.LocalAddr().Network(), ln.LocalAddr().String())
	require.NoError(t, err)

	var conn net.Conn
	retry.Run(t, func(t *retry.SubT) {
		conn, err = net.DialUDP("udp", nil, laddr)
		assert.NoError(t, err)
	})

	return srv, conn
}

func setTestHookServerServe(fn func(conn *net.UDPConn)) {
	testHookServerServe = fn
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
