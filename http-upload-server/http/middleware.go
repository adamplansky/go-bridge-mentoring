package http

import (
	"compress/gzip"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

func gzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func basicAuth(users *sync.Map) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			fmt.Println(users)
			fmt.Println(user, pass)

			credPass, credUserOk := users.Load(user)
			if !credUserOk {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			password, ok := credPass.(string)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if !credUserOk || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func rateLimiter(l *rate.Limiter) func(next http.Handler) http.Handler {
	limiter := l
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := limiter.Wait(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusRequestTimeout)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
