package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	_, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
	}

	http.StripPrefix("/files/", http.FileServer(http.Dir(h.uploadFolder))).ServeHTTP(w, r)
}
