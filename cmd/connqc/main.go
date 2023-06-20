package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/ettle/strcase"
	"github.com/hamba/cmd/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli/v2"
)

const (
	flagProtocol    = "protocol"
	flagProtocolTCP = "tcp"
	flagProtocolUDP = "udp"
	flagAddr        = "addr"

	flagBufferSize   = "buffer-size"
	flagReadTimeout  = "read-timeout"
	flagWriteTimeout = "write-timeout"

	flagConnBackoff  = "backoff"
	flagSendInterval = "interval"
)

var version = "¯\\_(ツ)_/¯"

var commands = []*cli.Command{
	{
		Name:  "client",
		Usage: "Run the connqc client",
		Flags: cmd.Flags{
			&cli.StringFlag{
				Name: flagProtocol,
				Usage: fmt.Sprintf(
					"The protocol for the connection. Supported protocols: '%s', '%s'", flagProtocolTCP, flagProtocolUDP,
				),
				Value:   flagProtocolTCP,
				EnvVars: []string{strcase.ToSNAKE(flagProtocol)},
			},
			&cli.StringFlag{
				Name:     flagAddr,
				Usage:    "The address of the connqc server",
				Required: true,
				EnvVars:  []string{strcase.ToSNAKE(flagAddr)},
			},
			&cli.DurationFlag{
				Name:    flagConnBackoff,
				Usage:   "The duration to wait for before retrying to connect to the server",
				Value:   time.Second,
				EnvVars: []string{strcase.ToSNAKE(flagConnBackoff)},
			},
			&cli.DurationFlag{
				Name:    flagSendInterval,
				Usage:   "The interval at which to send probe messages to the server",
				Value:   time.Second,
				EnvVars: []string{strcase.ToSNAKE(flagSendInterval)},
			},
			&cli.DurationFlag{
				Name:    flagReadTimeout,
				Usage:   "The duration after which the client should timeout when reading from a connection",
				Value:   2 * time.Second,
				EnvVars: []string{strcase.ToSNAKE(flagReadTimeout)},
			},
			&cli.DurationFlag{
				Name:    flagWriteTimeout,
				Usage:   "The duration after which the client should timeout when writing to a connection",
				Value:   5 * time.Second,
				EnvVars: []string{strcase.ToSNAKE(flagWriteTimeout)},
			},
		}.Merge(cmd.LogFlags),
		Action: runClient,
	},
	{
		Name:  "server",
		Usage: "Run the connqc server",
		Flags: cmd.Flags{
			&cli.StringFlag{
				Name:    flagAddr,
				Usage:   "The address to listen on for probe messages",
				Value:   ":8123",
				EnvVars: []string{strcase.ToSNAKE(flagAddr)},
			},
			&cli.IntFlag{
				Name:    flagBufferSize,
				Usage:   "The size of the read buffer used by the server",
				Value:   512,
				EnvVars: []string{strcase.ToSNAKE(flagBufferSize)},
			},
			&cli.DurationFlag{
				Name:    flagReadTimeout,
				Usage:   "The duration after which the server should timeout when reading from a connection",
				Value:   2 * time.Second,
				EnvVars: []string{strcase.ToSNAKE(flagReadTimeout)},
			},
			&cli.DurationFlag{
				Name:    flagWriteTimeout,
				Usage:   "The duration after which the server should timeout when writing to a connection",
				Value:   5 * time.Second,
				EnvVars: []string{strcase.ToSNAKE(flagWriteTimeout)},
			},
		}.Merge(cmd.LogFlags),
		Action: runServer,
	},
}

func main() {
	os.Exit(realMain())
}

func realMain() (code int) {
	ui := newTerm()

	defer func() {
		if v := recover(); v != nil {
			ui.Error(fmt.Sprintf("Panic: %v\n%s", v, string(debug.Stack())))
			code = 1
			return
		}
	}()

	app := cli.NewApp()
	app.Name = "connqc"
	app.Description = "Connection quality checker"
	app.Version = version
	app.Commands = commands

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := app.RunContext(ctx, os.Args); err != nil {
		ui.Error(err.Error())
		return 1
	}
	return 0
}
