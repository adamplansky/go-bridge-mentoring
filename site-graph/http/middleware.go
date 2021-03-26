package http

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"
)

const (
	// passwordDigst - hardcoded bcrypt password hash. password: "secret" in plaintext
	dbPasswordDigst = "$2a$04$QqzenEKqI.CNHTMvH0dPkeQqhLptBURwJSPlKFD0xt1QaQPN/rz26"
	dbUsername      = "gobridge"
)

var (
	ErrUsernameOrPasswordInvalid = fmt.Errorf("username or password is invalid")
)

// isAuthenticated return false if username or password
// is not allowed to access application.
func authenticate(username, password []byte) error {
	// ConstantTimeCompare returns 1 if the two slices, x and y, have equal contents
	// and 0 otherwise. The time taken is a function of the length of the slices and
	// is independent of the contents.
	if subtle.ConstantTimeCompare(username, []byte(dbUsername)) == 0 {
		return ErrUsernameOrPasswordInvalid
	}
	err := bcrypt.CompareHashAndPassword([]byte(dbPasswordDigst), password)
	switch err {
	case nil:
		// password match
		return nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return ErrUsernameOrPasswordInvalid
	default:
		return err
	}

}

func (s *server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Header)
		username, password, authOK := r.BasicAuth()
		fmt.Println(username, password, authOK)
		if authOK == false {
			httpErr(w, 401, ErrUnauthorized)
			return
		}

		if err := authenticate([]byte(username), []byte(password)); err != nil {
			if errors.Is(err, ErrUsernameOrPasswordInvalid) {
				http.StatusText(http.StatusUnauthorized)
				return
			}
			s.log.Error("authenticate failed", zap.Error(err))
			http.StatusText(http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
