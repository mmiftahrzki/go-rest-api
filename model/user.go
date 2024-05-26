package model

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        uuid.UUID `json:"id"`
	Email     string    `json:"email" validate:"required,email,max=100"`
	Password  string    `json:"password,omitempty" validate:"required,max=32"`
	Fullname  string    `json:"fullname" validate:"required,max=255"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by,omitempty"`
}

// type IUserModel interface {
// 	Create(ctx context.Context, id, username, email, fullname, gender, dob string) error
// 	FindAll(ctx context.Context, max_limit int) ([]User, error)
// 	FindById(ctx context.Context, id uuid.UUID) (User, error)
// 	FindAfterId(ctx context.Context, id string, max_limit int) ([]User, error)
// 	FindBeforeId(ctx context.Context, id string, max_limit int) ([]User, error)
// }

type userModel struct {
	database_connection *sql.DB
	table               string
	columns             []string
}

func NewUserModel(db *sql.DB, table string) *userModel {
	if db == nil || table == "" {
		log.Panicln("new user model: invalid params initial")
	}

	sql_query := fmt.Sprintf("SELECT * FROM %s LIMIT 1", table)
	rows, err := db.Query(sql_query)
	if err != nil {
		log.Panicln("new user model: ", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Panicln("new user model: ", err)
	}

	return &userModel{
		database_connection: db,
		table:               table,
		columns:             columns,
	}
}

func (model *userModel) Insert(ctx context.Context, user User) error {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return err
	}
	id := uuid.New()
	now := time.Now().In(loc)
	hmac_sha256 := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET_KEY")))
	hmac_sha256.Write([]byte(user.Password))

	password, err := bcrypt.GenerateFromPassword(hmac_sha256.Sum(nil), 12)
	if err != nil {
		log.Println(err)

		return err
	}

	sql_query := fmt.Sprintf(
		`INSERT INTO
			%s (%s)
		VALUES (
			unhex(replace(?, '-', '')),
			UPPER(?),
			?,
			?,
			?,
			?,
			?
		)`, model.table, strings.Join(model.columns, ", "))

	_, err = model.database_connection.ExecContext(ctx, sql_query, id, id.String(), user.Email, string(password), user.Fullname, now)
	if err != nil {
		log.Println(err)

		return err
	}

	return nil
}

func (model *userModel) FindAll(ctx context.Context, max_limit int) ([]User, error) {
	var users []User
	var User User
	var sql_query string
	var rows *sql.Rows
	var err error

	sql_query = fmt.Sprintf("SELECT %s FROM user ORDER BY fullname ASC LIMIT ?", strings.Join(model.columns, ", "))
	rows, err = model.database_connection.QueryContext(ctx, sql_query, max_limit+1)
	if err != nil {
		log.Println(err)

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		// err = rows.Scan(&User.Id, &User.Fullname, &User.Gender, &User.Email, &User.Username, &User.DateOfBirth, &User.CreatedAt)
		if err != nil {
			log.Println(err)

			return nil, err
		}

		users = append(users, User)
	}

	return users, nil
}

// func (model *userModel) FindAfterId(ctx context.Context, id string, max_limit int) ([]User, error) {
// 	var users []User
// 	var User User
// 	var sql_query string
// 	var rows *sql.Rows
// 	var err error

// 	sql_query = "SELECT fullname FROM user WHERE id_text = ?"
// 	if err = model.database_connection.QueryRowContext(ctx, sql_query, id).Scan(&User.Fullname); err != nil {
// 		log.Println(err)

// 		return nil, err
// 	}

// 	sql_query = fmt.Sprintf("SELECT %s FROM user WHERE fullname > ? ORDER BY fullname ASC LIMIT ?", model.fields)
// 	rows, err = model.database_connection.QueryContext(ctx, sql_query, User.Fullname, max_limit+1)
// 	if err != nil {
// 		log.Println(err)

// 		return nil, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		err = rows.Scan(&User.Id, &User.Fullname, &User.Gender, &User.Email, &User.Username, &User.DateOfBirth, &User.CreatedAt)
// 		if err != nil {
// 			log.Println(err)

// 			return nil, err
// 		}

// 		users = append(users, User)
// 	}

// 	return users, nil
// }

// func (model *userModel) FindBeforeId(ctx context.Context, id string, max_limit int) ([]User, error) {
// 	var users []User
// 	var User User
// 	var sql_query string
// 	var rows *sql.Rows
// 	var err error

// 	sql_query = "SELECT fullname FROM user WHERE id_text = ?"
// 	if err := model.database_connection.QueryRowContext(ctx, sql_query, id).Scan(&User.Fullname); err != nil {
// 		log.Println(err)

// 		return nil, err
// 	}

// 	sql_query = fmt.Sprintf(`
// 	SELECT a.* FROM (
// 		SELECT %s FROM user WHERE fullname < ? ORDER BY fullname DESC LIMIT ?
// 	) a
// 	ORDER BY a.fullname ASC;`, model.fields)
// 	rows, err = model.database_connection.QueryContext(ctx, sql_query, User.Fullname, max_limit+1)
// 	if err != nil {
// 		log.Println(err)

// 		return nil, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		err = rows.Scan(&User.Id, &User.Fullname, &User.Gender, &User.Email, &User.Username, &User.DateOfBirth, &User.CreatedAt)
// 		if err != nil {
// 			log.Println(err)

// 			return nil, err
// 		}

// 		users = append(users, User)
// 	}

// 	return users, nil
// }

// func (model *userModel) FindById(ctx context.Context, id uuid.UUID) (User, error) {
// 	var user User

// 	sql_query := fmt.Sprintf("SELECT %s FROM user WHERE id_text=?", model.fields)
// 	rows, err := model.database_connection.QueryContext(ctx, sql_query, id)
// 	if err != nil {
// 		log.Println(err)

// 		return user, err
// 	}
// 	defer rows.Close()

// 	if rows.Next() {
// 		err := rows.Scan(&user.Id, &user.Username, &user.Email, &user.Fullname, &user.Gender, &user.DateOfBirth, &user.CreatedAt)
// 		if err != nil {
// 			log.Println(err)

// 			return user, err
// 		}
// 	}

// 	return user, nil
// }
