package signal

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
)

// Client attempts to hold a connection with a server, sending probe messages at a configured interval.
type Client struct {
	backoff      time.Duration
	sendInterval time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration

	log *logger.Logger
}

// NewClient returns a client.
func NewClient(backoff, sendInterval, readTimeout, writeTimeout time.Duration, log *logger.Logger) *Client {
	return &Client{
		backoff:      backoff,
		sendInterval: sendInterval,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		log:          log,
	}
}

// Run sends probe messages to the server continuously.
// If the connection fails, it retries at the configured backoff interval.
func (c *Client) Run(ctx context.Context, addr string) {
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			c.log.Error("Could not connect to server", lctx.Str("address", addr))

			select {
			case <-ctx.Done():
				return
			case <-time.After(c.backoff):
				continue
			}
		}

		if err = c.handleConn(ctx, conn); err != nil {
			c.log.Error("Connection error", lctx.Error("error", err))
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

type expectation struct {
	timestamp time.Time
	probe     Probe
}

func (c *Client) handleConn(ctx context.Context, conn net.Conn) error { //nolint:funlen
	defer func() { _ = conn.Close() }()

	readCh := make(chan readResponse)
	go c.readLoop(conn, readCh)

	enc := NewEncoder(conn)

	id := uint64(1)
	var expect []expectation
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.sendInterval):
			_ = conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))

			p := Probe{
				ID:   id,
				Data: fmt.Sprintf("Hello %d", id),
			}
			if err := enc.Encode(p); err != nil {
				return fmt.Errorf("writing Message: %w", err)
			}

			c.log.Debug("Sent probe", lctx.Interface("probe", p))

			id++
			expect = append(expect, expectation{timestamp: time.Now(), probe: p})
		case resp, ok := <-readCh:
			if !ok {
				return nil
			}
			if resp.err != nil {
				return fmt.Errorf("reading Message: %w", resp.err)
			}

			var (
				exp   expectation
				found bool
			)
			for {
				if len(expect) == 0 {
					break
				}
				exp, expect = expect[0], expect[1:]
				if exp.probe.ID == resp.probe.ID {
					found = true
					break
				}

				c.log.Warn("Message lost", lctx.Uint64("id", exp.probe.ID), lctx.Str("data", exp.probe.Data))
			}
			if !found {
				c.log.Error("No expectation found")
				continue
			}

			c.log.Info("Message received",
				lctx.Uint64("id", exp.probe.ID),
				lctx.Str("data", exp.probe.Data),
				lctx.Duration("took", resp.timestamp.Sub(exp.timestamp)),
			)
		}
	}
}

type readResponse struct {
	timestamp time.Time
	probe     Probe
	err       error
}

func (c *Client) readLoop(conn net.Conn, ch chan readResponse) {
	defer close(ch)

	dec := NewDecoder(conn)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(c.readTimeout))

		msg, err := dec.Decode()
		if err != nil {
			ch <- readResponse{err: err}
			return
		}

		p, ok := msg.(Probe)
		if !ok {
			ch <- readResponse{err: err}
			continue
		}

		c.log.Debug("Received probe", lctx.Interface("probe", p))

		ch <- readResponse{timestamp: time.Now(), probe: p}
	}
}
