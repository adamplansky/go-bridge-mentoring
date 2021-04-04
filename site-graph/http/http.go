package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/crawler"
)

type Cache interface {
	// Add insert new key-value into cache, if key already exists
	// in cache returns evicted=true
	Add(key string, value interface{})
	// Get gets value from the coresponding key from cache
	// ok == true if object exist in cache, otherwire ok == false
	Get(key string) (value interface{}, ok bool)
}

type server struct {
	router  *mux.Router
	log     *zap.SugaredLogger
	crawler *crawler.Crawler
	cache   Cache
}

func NewServer(log *zap.SugaredLogger, c Cache) *server {
	s := server{
		log:     log,
		crawler: crawler.New(log, c),
		cache:   c,
	}
	s.router = s.routes()
	return &s
}

func (s *server) Run(ctx context.Context, addr string) error {
	httpServer := http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    15 * time.Second,           // the maximum duration for reading the entire request, including the body
		WriteTimeout:   15 * time.Second,           // the maximum duration before timing out writes of the response
		IdleTimeout:    30 * time.Second,           // the maximum amount of time to wait for the next request when keep-alive is enabled
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1 MB
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				s.log.Fatal("HTTP server ListenAndServe failed", zap.Error(err))
			}
		}
	}()
	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()
	return httpServer.Shutdown(shutdownCtx)
}
