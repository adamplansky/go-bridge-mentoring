package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/crawler"
)

type server struct {
	router  *mux.Router
	log     *zap.SugaredLogger
	crawler *crawler.Crawler
}

func NewServer(log *zap.SugaredLogger) *server {
	s := server{
		router:  mux.NewRouter(),
		log:     log,
		crawler: crawler.New(log),
	}
	s.routes()
	return &s
}

func (s *server) Run(ctx context.Context, addr string) error {
	httpServer := http.Server{
		Addr:              addr,
		Handler:           s.router,
		TLSConfig:         nil,
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ConnContext: nil,
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
