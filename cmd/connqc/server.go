package main

import (
	"errors"

	"github.com/hamba/cmd/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/nitrado/connqc"
	"github.com/nitrado/connqc/tcp"
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

	sigSrv := connqc.NewServer(bufferSize, readTimeout, writeTimeout, log)

	srv, err := tcp.NewServer(sigSrv)
	if err != nil {
		return err
	}

	addr := c.String(flagAddr)
	log.Info("Starting server",
		lctx.Str("addr", addr),
		lctx.Int("buffer_size", bufferSize),
		lctx.Duration("read_timeout", readTimeout),
		lctx.Duration("write_timeout", writeTimeout),
	)
	go func() {
		if err := srv.Listen(addr); err != nil && !errors.Is(err, tcp.ErrServerClosed) {
			log.Error("Server error", lctx.Error("error", err))
		}
	}()
	defer func() { _ = srv.Close() }()

	<-ctx.Done()

	log.Info("Shutting down")

	if err = srv.Close(); err != nil {
		log.Warn("Failed to shutdown server", lctx.Error("error", err))
	}

	return nil
}
