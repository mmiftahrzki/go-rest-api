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
	"github.com/mmiftahrzki/go-rest-api/middleware"
	"github.com/mmiftahrzki/go-rest-api/response"
	"github.com/mmiftahrzki/go-rest-api/router"
)

type JwtClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type signinpayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type jwtContextKey int

const key jwtContextKey = iota
const req_header_auth_key string = "Authorization"

var errEmptyAuth = errors.New("authorization header not found")
var errInvalidAuth = errors.New("invalid authorization header")

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

func New() middleware.Middleware {
	return authHandler
}

func authHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		auth_value := request.Header.Get(req_header_auth_key)
		token_str, err := extractAuthTokenStr(auth_value)
		if err != nil {
			router.Response.Message = err.Error()

			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(router.Response.ToJson()))

			return
		}

		token, err := jwt.ParseWithClaims(token_str, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			method, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok || method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("invalid signing method")
			}

			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			router.Response.Message = err.Error()

			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(router.Response.ToJson()))

			return
		}

		if !token.Valid {
			router.Response.Message = "invalid jwt"

			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte(router.Response.ToJson()))

			return
		}

		claims, ok := token.Claims.(*JwtClaims)
		if !ok {
			router.Response.Message = "invalid jwt claims"

			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(router.Response.ToJson()))

			return
		}

		request = request.WithContext(context.WithValue(request.Context(), key, claims))

		next.ServeHTTP(writer, request)
	}
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

	// registerd_claims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute))}
	// claims := JwtClaims{
	// 	Email:            payload.Email,
	// 	RegisteredClaims: registerd_claims,
	// }

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// ss, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	token, err := GenerateToken(payload)
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(response.ToJson())

		return
	}

	response.Data["token"] = token
	response.Message = "berhasil generate token"

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(response.ToJson())
}

func GenerateToken(payload signinpayload) (string, error) {
	registerd_claims := jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute))}
	claims := JwtClaims{
		Email:            payload.Email,
		RegisteredClaims: registerd_claims,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed_string, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return signed_string, err
	}

	return signed_string, nil
}
