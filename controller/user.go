package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/database"
	"github.com/mmiftahrzki/go-rest-api/model"
	"github.com/mmiftahrzki/go-rest-api/response"
	"golang.org/x/crypto/bcrypt"
)

type userLoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// type controller struct {
// 	model model.User
// }

// func (c *controller) CreateUser(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
func CreateUser(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	// marshal http request body payload to user type struct
	request_body, err := io.ReadAll(request.Body)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	user := &model.User{}
	err = json.Unmarshal(request_body, user)
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	// validate user struct type
	validator := validator.New()
	err = validator.Struct(user)
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	id := uuid.New()
	now := time.Now().In(loc)
	hmac_sha256 := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET_KEY")))
	hmac_sha256.Write([]byte(user.Password))

	password_hash, err := bcrypt.GenerateFromPassword(hmac_sha256.Sum(nil), 12)
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)

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

		mysql_error, ok := err.(*mysql.MySQLError)
		if ok {
			if mysql_error.Number == 1062 {
				res.Message = fmt.Sprintf("user dengan email: %s sudah ada", user.Email)

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusConflict)
				writer.Write(res.ToJson())

				return
			}
		}

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	res.Message = "berhasil membuat user baru"
	res.Data["id"] = id.String()

	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(res.ToJson()))
}

func ReadUser(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	// marshal http request body payload to user login payload type struct
	request_body, err := io.ReadAll(request.Body)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	user_login := &userLoginPayload{}
	err = json.Unmarshal(request_body, user_login)
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	hmac_sha256 := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET_KEY")))
	hmac_sha256.Write([]byte(user_login.Password))

	sql_query := "SELECT id_text, email, password, fullname, created_at FROM user WHERE email=?;"
	db := database.GetDatabaseConnection()
	rows := db.QueryRowContext(request.Context(), sql_query, user_login.Email)
	err = rows.Err()
	if err != nil {
		log.Println(err)

		return
	}

	user := &model.User{}
	var stored_hashed_password []byte
	err = rows.Scan(&user.Id, &user.Email, &stored_hashed_password, &user.Fullname, &user.CreatedAt)
	if err != nil {
		log.Println(err)

		return
	}

	err = bcrypt.CompareHashAndPassword(stored_hashed_password, hmac_sha256.Sum(nil))
	if err != nil {
		log.Println(err)

		return
	}

	res.Message = "berhasil mengambil data user"
	res.Data["user"] = user

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(res.ToJson()))
}
