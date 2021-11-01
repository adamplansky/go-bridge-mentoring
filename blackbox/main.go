package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/adamplansky/go-bridge-mentoring/blackbox/blackbox"
)

func run() error {
	//sig := make(chan os.Signal, 1)
	////signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	jobs := blackbox.ParseInput()
	logReporter, err := blackbox.NewLogResult()
	if err != nil {
		return err
	}
	s, err := blackbox.NewScraper(jobs, logReporter)
	if err != nil {
		return err
	}

	resultCh := s.RunReporter()
	defer close(resultCh)

	s.Scrape(ctx, resultCh)

	<-ctx.Done()

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
