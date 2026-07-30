package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/app"
)

// Entry point
func main() {
	os.Exit(runWithSignals(os.Args[1:], os.Stdout, os.Stderr))
}

func runWithSignals(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return run(ctx, args, stdout, stderr)
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	cfg, err := parseConfig(args)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	application, err := app.New(cfg)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	if err := application.Run(ctx, stdout); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
