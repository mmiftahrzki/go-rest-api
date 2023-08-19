package model

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Fullname    string    `json:"fullname"`
	Gender      string    `json:"gender"`
	DateOfBirth time.Time `json:"date_of_birth"`
	CreatedAt   time.Time `json:"created_at"`
}

type IUserModel interface {
	Create(ctx context.Context, id, username, email, fullname, gender, dob string) error
	FindAll(ctx context.Context, max_limit int) ([]User, error)
	FindById(ctx context.Context, id uuid.UUID) (User, error)
	FindAfterId(ctx context.Context, id string, max_limit int) ([]User, error)
	FindBeforeId(ctx context.Context, id string, max_limit int) ([]User, error)
}

type userModel struct {
	database_connection *sql.DB
	fields              string
}

func NewUserModel(db *sql.DB) IUserModel {
	return &userModel{
		database_connection: db,
		fields:              "id_text, fullname, gender, email, username, date_of_birth, created_at",
	}
}

func (model *userModel) Create(ctx context.Context, id, username, email, fullname, gender, dob string) error {
	sql_query :=
		`INSERT INTO
			user(id, username, email, fullname, gender, date_of_birth)
		VALUES (unhex(replace(?, '-', '')), ?, ?, ?, ?, ?)`

	_, err := model.database_connection.ExecContext(ctx, sql_query, id, username, email, fullname, gender, dob)

	return err
}

func (model *userModel) FindAll(ctx context.Context, max_limit int) ([]User, error) {
	var users []User
	var User User
	var sql_query string
	var rows *sql.Rows
	var err error

	sql_query = fmt.Sprintf("SELECT %s FROM user ORDER BY fullname ASC LIMIT ?", model.fields)
	rows, err = model.database_connection.QueryContext(ctx, sql_query, max_limit+1)
	if err != nil {
		log.Println(err)

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&User.Id, &User.Fullname, &User.Gender, &User.Email, &User.Username, &User.DateOfBirth, &User.CreatedAt)
		if err != nil {
			log.Println(err)

			return nil, err
		}

		users = append(users, User)
	}

	return users, nil
}

func (model *userModel) FindAfterId(ctx context.Context, id string, max_limit int) ([]User, error) {
	var users []User
	var User User
	var sql_query string
	var rows *sql.Rows
	var err error

	sql_query = "SELECT fullname FROM user WHERE id_text = ?"
	if err = model.database_connection.QueryRowContext(ctx, sql_query, id).Scan(&User.Fullname); err != nil {
		log.Println(err)

		return nil, err
	}

	sql_query = fmt.Sprintf("SELECT %s FROM user WHERE fullname > ? ORDER BY fullname ASC LIMIT ?", model.fields)
	rows, err = model.database_connection.QueryContext(ctx, sql_query, User.Fullname, max_limit+1)
	if err != nil {
		log.Println(err)

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&User.Id, &User.Fullname, &User.Gender, &User.Email, &User.Username, &User.DateOfBirth, &User.CreatedAt)
		if err != nil {
			log.Println(err)

			return nil, err
		}

		users = append(users, User)
	}

	return users, nil
}

func (model *userModel) FindBeforeId(ctx context.Context, id string, max_limit int) ([]User, error) {
	var users []User
	var User User
	var sql_query string
	var rows *sql.Rows
	var err error

	sql_query = "SELECT fullname FROM user WHERE id_text = ?"
	if err := model.database_connection.QueryRowContext(ctx, sql_query, id).Scan(&User.Fullname); err != nil {
		log.Println(err)

		return nil, err
	}

	sql_query = fmt.Sprintf(`
	SELECT a.* FROM (
		SELECT %s FROM user WHERE fullname < ? ORDER BY fullname DESC LIMIT ?
	) a
	ORDER BY a.fullname ASC;`, model.fields)
	rows, err = model.database_connection.QueryContext(ctx, sql_query, User.Fullname, max_limit+1)
	if err != nil {
		log.Println(err)

		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&User.Id, &User.Fullname, &User.Gender, &User.Email, &User.Username, &User.DateOfBirth, &User.CreatedAt)
		if err != nil {
			log.Println(err)

			return nil, err
		}

		users = append(users, User)
	}

	return users, nil
}

func (model *userModel) FindById(ctx context.Context, id uuid.UUID) (User, error) {
	var user User

	sql_query := fmt.Sprintf("SELECT %s FROM user WHERE id_text=?", model.fields)
	rows, err := model.database_connection.QueryContext(ctx, sql_query, id)
	if err != nil {
		log.Println(err)

		return user, err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&user.Id, &user.Username, &user.Email, &user.Fullname, &user.Gender, &user.DateOfBirth, &user.CreatedAt)
		if err != nil {
			log.Println(err)

			return user, err
		}
	}

	return user, nil
}
