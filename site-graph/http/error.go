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

func HttpError(w http.ResponseWriter, code int, err error) error {
	e := &MyError{
		HTTPStatus: code,
		Message:    err.Error(),
	}

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	w.WriteHeader(e.HTTPStatus)
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	return nil
}
