package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"go.uber.org/zap"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/cache/cachemap"
	"github.com/adamplansky/go-bridge-mentoring/site-graph/http"
)

func main() {
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	defer logger.Sync() // flushes buffer, if any

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cache := cachemap.New()
	server := http.NewServer(log, cache)
	fmt.Println("I'm alive :8080")
	if err := server.Run(ctx, ":8080"); err != nil {
		log.Fatal(err)
	}
}
