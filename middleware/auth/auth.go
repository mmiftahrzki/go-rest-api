package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/response"
	"github.com/mmiftahrzki/go-rest-api/router"
)

type JwtClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type jwtContextKey int

const key jwtContextKey = iota
const req_header_auth_key string = "Authorization"

var errEmptyAuth = errors.New("auth: authorization header not found")
var errInvalidAuth = errors.New("auth: invalid authorization header")

func extractAuthTokenStr(auth_value string) (string, error) {
	var token_str string

	if len(auth_value) == 0 {
		return token_str, errEmptyAuth
	}

	auth_value_fields := strings.Fields(auth_value)
	if len(auth_value_fields) != 2 || auth_value_fields[0] != "Bearer" {
		return token_str, errInvalidAuth
	}

	token_str = auth_value_fields[1]

	return token_str, nil
}

func ExtractAuthClaims(ctx context.Context) (*JwtClaims, error) {
	auth_value := ctx.Value(key)
	claims, ok := auth_value.(*JwtClaims)
	if !ok {
		return nil, errors.New("auth: invalid jwt claims")
	}

	return claims, nil
}

func New() router.Middleware {
	return authHandler
}

func authHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		auth_value := request.Header.Get(req_header_auth_key)
		token_str, err := extractAuthTokenStr(auth_value)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		token, err := jwt.ParseWithClaims(token_str, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			method, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok || method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("auth: invalid signing method")
			}

			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		claims, ok := token.Claims.(*JwtClaims)
		if !ok || !token.Valid {
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		request = request.WithContext(context.WithValue(request.Context(), key, claims))

		next.ServeHTTP(writer, request)
	}
}

type signinpayload struct {
	Email string
	// password string
}

func NewSignInPayload() *signinpayload {
	return &signinpayload{}
}

func Token(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	var payload signinpayload
	response := response.New()

	json_decoder := json.NewDecoder(request.Body)
	err := json_decoder.Decode(&payload)
	if err != nil {
		log.Println(err)

		response.Message = http.StatusText(http.StatusBadRequest)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(response.ToJson())

		return
	}

	registerd_claims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute))}
	claims := JwtClaims{
		Email:            payload.Email,
		RegisteredClaims: registerd_claims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(response.ToJson())

		return
	}

	response.Data["token"] = ss
	response.Message = "berhasil generate token"

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(200)
	writer.Write(response.ToJson())
}
