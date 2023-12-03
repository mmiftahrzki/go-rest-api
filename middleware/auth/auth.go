package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mmiftahrzki/go-rest-api/middleware"
)

// type middleware struct {
// 	next http.Handler
// 	w    http.ResponseWriter
// 	r    *http.Request
// }

type JwtClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var ErrEmptyAuth = errors.New("authorization header not found")
var ErrInvalidAuth = errors.New("invalid authorization header")
var Claims *JwtClaims

const auth_key = "Authorization"

func extractAuthTokenStr(r *http.Request) (string, error) {
	var token_str string

	auth_value := r.Header.Get(auth_key)
	if len(auth_value) == 0 {
		return token_str, ErrEmptyAuth
	}

	auth_value_fields := strings.Fields(auth_value)
	if len(auth_value_fields) != 2 || auth_value_fields[0] != "Bearer" {
		return token_str, ErrInvalidAuth
	}

	token_str = auth_value_fields[1]

	return token_str, nil
}

func New() middleware.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		content_type := r.Header.Get("Content-Type")
		if content_type != "application/json" {
			w.WriteHeader(http.StatusBadRequest)

			return errors.New(http.StatusText(http.StatusBadRequest))
		}

		token_str, err := extractAuthTokenStr(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return errors.New(err.Error())
		}

		token, err := jwt.ParseWithClaims(token_str, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			method, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok || method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("invalid signing method")
			}

			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			http_err_status_code := http.StatusInternalServerError

			err, ok := err.(*jwt.ValidationError)
			if ok {
				http_err_status_code = http.StatusBadRequest
			}

			w.WriteHeader(http_err_status_code)

			return errors.New(err.Error())
		}

		claims, ok := token.Claims.(*JwtClaims)
		if !ok || !token.Valid {
			w.WriteHeader(http.StatusBadRequest)

			return errors.New(err.Error())
		}

		Claims = claims

		return nil
	}
}

// func NewAuth(router *httprouter.Router) *middleware {
// 	return &middleware{next: router}
// }

// func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	content_type := r.Header.Get("Content-Type")
// 	if content_type != "application/json" {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte(http.StatusText(http.StatusBadRequest)))

// 		return
// 	}

// 	token_str, err := extractAuthTokenStr(r)
// 	if err != nil {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		w.Write([]byte(err.Error()))

// 		return
// 	}

// 	token, err := jwt.ParseWithClaims(token_str, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
// 		method, ok := t.Method.(*jwt.SigningMethodHMAC)
// 		if !ok || method != jwt.SigningMethodHS256 {
// 			return nil, fmt.Errorf("signing method invalid")
// 		}

// 		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
// 	})

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte(err.Error()))

// 		return
// 	}

// 	claims, ok := token.Claims.(*JwtClaims)
// 	if !ok || !token.Valid {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte(err.Error()))

// 		return
// 	}

// 	Claims = claims
// 	m.next.ServeHTTP(w, r)
// }
