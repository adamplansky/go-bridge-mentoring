package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

var (
	ErrUnauthorized = fmt.Errorf("not authorized")
	ErrInternal     = fmt.Errorf("internal server error")
)

// httpErr set status code and content-type into http header
// if error != nil it also create unified http error message
// and encode it into http.body
func (s *server) httpErr(w http.ResponseWriter, code int, err error) {
	type myError struct {
		Message string `json:"message"`
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		e := &myError{Message: err.Error()}
		// forcing shadowing err
		if err := json.NewEncoder(w).Encode(e); err != nil {
			s.log.Error("internal server error", zap.Error(err))
			http.Error(w, ErrInternal.Error(), 500)
			return
		}
		s.log.Warn("unable to process request", zap.Error(err))
		return
	}
}
