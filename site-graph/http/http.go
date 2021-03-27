package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/cache"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/crawler"
)

type server struct {
	router  *mux.Router
	log     *zap.SugaredLogger
	crawler *crawler.Crawler
	cache   cache.Cache
}

func NewServer(log *zap.SugaredLogger, c cache.Cache) *server {
	s := server{
		router:  mux.NewRouter(),
		log:     log,
		crawler: crawler.New(log, c),
		cache:   c,
	}
	s.routes()
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
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			s.log.Fatal("HTTP server ListenAndServe failed", zap.Error(err))
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
