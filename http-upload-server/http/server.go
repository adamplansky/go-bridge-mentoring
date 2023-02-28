package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

type server struct {
	router *chi.Mux
	log    *zap.Logger
}

func NewServer(log *zap.Logger, maxUploadRequests int) *server {
	return &server{
		log:    log,
		router: router(log, maxUploadRequests),
	}
}

func (s *server) Run(ctx context.Context, addr string) error {
	httpServer := http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    15 * time.Second,           // the maximum duration for reading the entire request, including the body
		WriteTimeout:   20 * time.Second,           // the maximum duration before timing out writes of the response
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
