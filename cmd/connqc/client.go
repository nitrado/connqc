package main

import (
	"context"
	"fmt"

	"github.com/hamba/cmd/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/nitrado/connqc"
	"github.com/urfave/cli/v2"
)

func runClient(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	log, err := cmd.NewLogger(c)
	if err != nil {
		return err
	}

	protocol := c.String(flagProtocol)
	if protocol != flagProtocolTCP && protocol != flagProtocolUDP {
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	backoff := c.Duration(flagConnBackoff)
	sendInterval := c.Duration(flagSendInterval)
	readTimeout := c.Duration(flagReadTimeout)
	writeTimeout := c.Duration(flagWriteTimeout)

	log = log.With(lctx.Str("protocol", protocol))

	client := connqc.NewClient(backoff, sendInterval, readTimeout, writeTimeout, log)
	go func() {
		client.Run(ctx, protocol, c.String(flagAddr))
		cancel()
	}()

	<-ctx.Done()

	log.Info("Shutting down")

	return nil
}
