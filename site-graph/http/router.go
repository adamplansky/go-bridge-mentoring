package http

import (
	"net/http"

	"github.com/adamplansky/go-bridge-mentoring/site-graph/http/static"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()

	//fileServer := http.FileServer(http.Dir("./http/static")) // New code
	fileServer := http.FileServer(http.FS(static.WebUI)) // New code
	r.Handle("/", fileServer)

	// v1 must contain basic auth
	s := r.PathPrefix("/v1").Subrouter()
	s.Use(authMiddleware)
	s.HandleFunc("/graph", ScrapeHandler()).Methods(http.MethodGet)

	return r
}
