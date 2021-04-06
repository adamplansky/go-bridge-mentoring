package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/http/static"
)

func (s *server) routes() *mux.Router {
	r := mux.NewRouter()
	fileServer := http.FileServer(http.FS(static.WebUI)) // New code
	r.Handle("/", fileServer)

	// v1 must contain basic auth
	sr := r.PathPrefix("/v1").Subrouter()
	sr.Use(s.authMiddleware)
	sr.Use(s.gzipHandler)
	sr.HandleFunc("/graph", s.ScrapeHandler).Methods(http.MethodGet)

	return r
}
