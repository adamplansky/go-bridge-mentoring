package http

import (
	"net/http"
)

// IsAuth return false if username or password
// is not allowed to access application.
func IsAuth(username, password string) bool {
	return username == "gobridge" && password == "secret"
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, authOK := r.BasicAuth()
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
