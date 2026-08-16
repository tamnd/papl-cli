// Command papl is a single-binary CLI for the PAPL textbook.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/tamnd/papl-cli/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cli.Root()
	root.SetContext(ctx)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
