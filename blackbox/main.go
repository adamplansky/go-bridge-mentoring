package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/adamplansky/go-bridge-mentoring/blackbox/blackbox"
)

func run() error {
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
	s.Scrape(resultCh)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}
