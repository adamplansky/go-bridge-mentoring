package http

import (
	"net/http"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/http/static"
)

func (s *server) routes() {
	fileServer := http.FileServer(http.FS(static.WebUI)) // New code
	s.router.Handle("/", fileServer)

	// v1 must contain basic auth
	sr := s.router.PathPrefix("/v1").Subrouter()
	sr.Use(s.authMiddleware)
	sr.Use(s.gzipHandler)
	sr.HandleFunc("/graph", s.ScrapeHandler).Methods(http.MethodGet)
}
