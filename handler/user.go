package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/database"
	"github.com/mmiftahrzki/go-rest-api/middleware/auth"
	"github.com/mmiftahrzki/go-rest-api/model"
	"github.com/mmiftahrzki/go-rest-api/router"
	"golang.org/x/crypto/bcrypt"
)

// func (c *controller) CreateUser(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
func CreateUser(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	writer.Header().Set("Content-Type", "application/json")

	// marshal http request body payload to user type struct
	request_body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Println(err)

		router.Response.Message = "terjadi kesalahan tak terduga di server. silakan coba lagi nanti."

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(router.Response.ToJson())

		return
	}

	user := &model.User{}
	buffer := strings.NewReader(string(request_body))
	json_decoder := json.NewDecoder(buffer)
	err = json_decoder.Decode(user)
	if err != nil {
		log.Println(err)

		router.Response.Message = "invalid payload"

		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(router.Response.ToJson())

		return
	}

	// validate user struct type
	validator := validator.New()
	err = validator.Struct(user)
	if err != nil {
		log.Println(err)

		router.Response.Message = "invalid payload"

		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(router.Response.ToJson())

		return
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	id := uuid.New()
	now := time.Now().In(loc)
	hmac_sha256 := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET_KEY")))
	hmac_sha256.Write([]byte(user.Password))

	password_hash, err := bcrypt.GenerateFromPassword(hmac_sha256.Sum(nil), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)

		router.Response.Message = "terjadi kesalahan tak terduga di server. silakan coba lagi nanti."

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(router.Response.ToJson())

		return
	}

	sql_query :=
		`INSERT INTO
			user (
				id,
				id_text,
				email,
				password,
				fullname,
				created_at
			)
		VALUES (
			unhex(replace(?, '-', '')),
			UPPER(?),
			?,
			?,
			?,
			?
		);`

	db := database.GetDatabaseConnection()

	_, err = db.ExecContext(request.Context(), sql_query, id, id.String(), user.Email, string(password_hash), user.Fullname, now)
	if err != nil {
		log.Println(err)

		response_status := http.StatusInternalServerError
		response_message := "terjadi kesalahan tak terduga di server. silakan coba lagi nanti."

		mysql_error, ok := err.(*mysql.MySQLError)
		if ok {
			if mysql_error.Number == 1062 {
				response_status = http.StatusConflict
				response_message = fmt.Sprintf("user dengan email: %s sudah ada", user.Email)
			}
		}

		router.Response.Message = response_message

		writer.WriteHeader(response_status)
		writer.Write(router.Response.ToJson())

		return
	}

	router.Response.Message = "berhasil membuat user baru"
	router.Response.Data["id"] = id.String()

	writer.WriteHeader(http.StatusCreated)
	writer.Write(router.Response.ToJson())
}

func ReadUser(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	writer.Header().Set("Content-Type", "application/json")

	// marshal http request body payload to user login payload type struct
	request_body, err := io.ReadAll(request.Body)
	if err != nil {
		log.Println(err)

		router.Response.Message = "terjadi kesalahan tak terduga di server. silakan coba lagi nanti."

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(router.Response.ToJson())

		return
	}

	user_login := auth.NewSignInPayload()
	buffer := bytes.NewBuffer(request_body)
	json_decoder := json.NewDecoder(buffer)
	err = json_decoder.Decode(user_login)
	if err != nil {
		log.Println(err)

		router.Response.Message = "invalid payload"

		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(router.Response.ToJson())

		return
	}

	hmac_sha256 := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET_KEY")))
	hmac_sha256.Write([]byte(user_login.Password))

	sql_query := "SELECT password FROM user WHERE email=?;"
	db := database.GetDatabaseConnection()
	row := db.QueryRowContext(request.Context(), sql_query, user_login.Email)

	var stored_hashed_password []byte
	err = row.Scan(&stored_hashed_password)
	if err != nil {
		log.Println(err)

		router.Response.Message = "terjadi kesalahan tak terduga di server. silakan coba lagi nanti."

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(router.Response.ToJson())

		return
	}

	err = bcrypt.CompareHashAndPassword(stored_hashed_password, hmac_sha256.Sum(nil))
	if err != nil {
		log.Println(err)

		router.Response.Message = "email atau password invalid"

		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write(router.Response.ToJson())

		return
	}

	token, err := auth.GenerateToken(*user_login)
	if err != nil {
		log.Println(err)

		router.Response.Message = "terjadi kesalahan tak terduga di server. silakan coba lagi nanti."

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(router.Response.ToJson())

		return
	}

	router.Response.Data["token"] = token
	router.Response.Message = "berhasil generate token"

	writer.WriteHeader(http.StatusOK)
	writer.Write(router.Response.ToJson())
}
