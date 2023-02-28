package http

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"go.uber.org/zap"

	"github.com/adamplansky/go-bridge-mentoring/http-upload-server/rest"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func users() *sync.Map {
	m := &sync.Map{}
	m.Store("user1", "password")
	m.Store("user2", "secret-password")
	return m
}

func router(logger *zap.Logger, maxConcurrentRequests int) *chi.Mux {
	r := chi.NewRouter()
	r.Use(Logger(logger))
	r.Use(gzipHandler)
	r.Use(basicAuth(users()))

	limiter := rate.NewLimiter(rate.Limit(maxConcurrentRequests), 1)
	r.Use(rateLimiter(limiter))

	h := rest.NewHandlers(logger)
	r.Route("/", func(r chi.Router) {
		r.Post("/upload", h.Upload)
		r.Get("/files/{id}", h.Download)
	})

	return r
}

func Logger(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				l.Info("Served",
					zap.String("proto", r.Proto),
					zap.String("path", r.URL.Path),
					zap.Duration("lat", time.Since(t1)),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
					zap.String("reqId", middleware.GetReqID(r.Context())))
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
