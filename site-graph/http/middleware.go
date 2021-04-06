package http

import (
	"compress/gzip"
	"crypto/subtle"
	"errors"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameOrPasswordInvalid = errors.New("username or password is invalid")
)

// isAuthenticated return false if username or password
// is not allowed to access application.
func authenticate(username, password []byte) error {
	var (
		// passwordDigst - hardcoded bcrypt password hash. password: "secret" in plaintext
		dbPasswordDigst = []byte("$2a$04$QqzenEKqI.CNHTMvH0dPkeQqhLptBURwJSPlKFD0xt1QaQPN/rz26")
		dbUsername      = []byte("gobridge")
	)

	// ConstantTimeCompare returns 1 if the two slices, x and y, have equal contents
	// and 0 otherwise. The time taken is a function of the length of the slices and
	// is independent of the contents.
	if subtle.ConstantTimeCompare(username, dbUsername) == 0 {
		return ErrUsernameOrPasswordInvalid
	}
	err := bcrypt.CompareHashAndPassword(dbPasswordDigst, password)
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return ErrUsernameOrPasswordInvalid
	}
	return err

}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, authOK := r.BasicAuth()
		if authOK == false {
			s.httpErr(w, http.StatusUnauthorized, nil)
			return
		}

		if err := authenticate([]byte(username), []byte(password)); err != nil {
			if errors.Is(err, ErrUsernameOrPasswordInvalid) {
				s.httpErr(w, http.StatusUnauthorized, err)
				return
			}
			s.log.Error("authenticate failed", zap.Error(err))
			s.httpErr(w, http.StatusInternalServerError, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *server) gzipHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

//
//func (s *server) gzipMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Encoding", "gzip")
//		gz := gzip.NewWriter(w)
//		defer gz.Close()
//		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
//
//		next.ServeHTTP(gzw, r)
//	})
//}
