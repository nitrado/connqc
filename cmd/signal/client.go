package main

import (
	"github.com/hamba/cmd/v2"
	"github.com/nitrado/connqc"
	"github.com/urfave/cli/v2"
)

func runClient(c *cli.Context) error {
	ctx := c.Context

	log, err := cmd.NewLogger(c)
	if err != nil {
		return err
	}

	backoff := c.Duration(flagConnBackoff)
	sendInterval := c.Duration(flagSendInterval)
	readTimeout := c.Duration(flagReadTimeout)
	writeTimeout := c.Duration(flagWriteTimeout)

	sig := connqc.NewClient(backoff, sendInterval, readTimeout, writeTimeout, log)

	addr := c.String(flagAddr)
	go sig.Run(ctx, addr)

	<-ctx.Done()

	log.Info("Shutting down")

	return nil
}
