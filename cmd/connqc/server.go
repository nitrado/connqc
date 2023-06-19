package main

import (
	"errors"
	"net"
	"sync"

	"github.com/hamba/cmd/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/nitrado/connqc"
	"github.com/nitrado/connqc/tcp"
	"github.com/nitrado/connqc/udp"
	"github.com/urfave/cli/v2"
)

func runServer(c *cli.Context) error {
	ctx := c.Context

	log, err := cmd.NewLogger(c)
	if err != nil {
		return err
	}

	bufferSize := c.Int(flagBufferSize)
	readTimeout := c.Duration(flagReadTimeout)
	writeTimeout := c.Duration(flagWriteTimeout)

	srv := connqc.NewServer(bufferSize, readTimeout, writeTimeout, log)

	tcpSrv, err := tcp.NewServer(srv)
	if err != nil {
		return err
	}

	udpSrv, err := udp.NewServer(srv)
	if err != nil {
		return err
	}

	addr := c.String(flagAddr)

	grp := sync.WaitGroup{}
	grp.Add(2)

	log.Info("Starting server",
		lctx.Str("addr", addr),
		lctx.Int("buffer_size", bufferSize),
		lctx.Duration("read_timeout", readTimeout),
		lctx.Duration("write_timeout", writeTimeout),
	)
	go func() {
		defer grp.Done()

		if err := tcpSrv.Listen(ctx, addr); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Error("Server error", lctx.Str("protocol", "tcp"), lctx.Err(err))
			return
		}
		log.Info("TCP server stopped")
	}()
	go func() {
		defer grp.Done()

		if err := udpSrv.Listen(ctx, addr); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Error("Server error", lctx.Str("protocol", "udp"), lctx.Err(err))
			return
		}
		log.Info("UDP server stopped")
	}()

	<-ctx.Done()

	log.Info("Shutting down")

	grp.Wait()

	return nil
}
