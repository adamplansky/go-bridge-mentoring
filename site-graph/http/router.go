package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.Use(authMiddleware)

	r.HandleFunc("/v1/graph", ScrapeHandler()).Methods(http.MethodGet)
	return r
}
