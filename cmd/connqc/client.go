package main

import (
	"fmt"

	"github.com/hamba/cmd/v2"
	lctx "github.com/hamba/logger/v2/ctx"
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

	var protocol string
	switch {
	case c.Bool(flagProtocolTCP) && !c.Bool(flagProtocolUDP):
		protocol = "tcp"
	case c.Bool(flagProtocolUDP) && !c.Bool(flagProtocolTCP):
		protocol = "udp"
	default:
		return fmt.Errorf("either --%s or --%s must be set", flagProtocolTCP, flagProtocolUDP)
	}
	log = log.With(lctx.Str("protocol", protocol))

	client := connqc.NewClient(backoff, sendInterval, readTimeout, writeTimeout, log)
	go client.Run(ctx, protocol, c.String(flagAddr))

	<-ctx.Done()

	log.Info("Shutting down")

	return nil
}
