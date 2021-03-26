package http

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	ErrUnauthorized = fmt.Errorf("not authorized")
)

type MyError struct {
	HTTPStatus int    `json:"-"`
	Message    string `json:"message"`
}

func httpErr(w http.ResponseWriter, code int, err error) {
	e := &MyError{
		HTTPStatus: code,
		Message:    err.Error(),
	}

	err = json.NewEncoder(w).Encode(e)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
