package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/http"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	h := http.Router()
	fmt.Println("I'm alive :8080")
	http.ServeHTTP(ctx, ":8080", h)
}
