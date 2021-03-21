package http

import (
	"fmt"
	"net/http"
)

// IsAuth return false if username or password
// is not allowed to access application.
func IsAuth(username, password string) bool {
	fmt.Println(username, password)
	return username == "gobridge" && password == "secret"
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header)
		username, password, authOK := r.BasicAuth()
		fmt.Println(username, password, authOK)
		if authOK == false {
			_ = HttpError(w, 401, ErrUnauthorized)
			return
		}

		if !IsAuth(username, password) {
			_ = HttpError(w, 401, ErrUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
