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

type MyError struct {
	Message string `json:"message"`
}

// httpErr set status code and content-type into http header
// if error != nil it also create unified http error message
// and encode it into http.body
func (s *server) httpErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		e := &MyError{Message: err.Error()}
		err = json.NewEncoder(w).Encode(e)
		if err == nil {
			code = http.StatusInternalServerError
		}
	}

	// if status code ==  500 always return unified error message
	if code == http.StatusInternalServerError {
		s.log.Error("internal server error", zap.Error(err))
		http.Error(w, ErrInternal.Error(), 500)
		return
	}

	s.log.Warn("unable to process request", zap.Error(err))
}
