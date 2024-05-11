package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mmiftahrzki/go-rest-api/middleware/auth"
)

const Max_limit int = 10

type Date time.Time

func (j *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if string(s) == "null" || string(s) == "" {
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}

	*j = Date(t)

	return nil
}

func (j Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(j))
}

func (j Date) Format() string {
	t := time.Time(j)

	return t.Format("2006-01-02")
}

type Customer struct {
	Id          uuid.UUID `json:"id"`
	Username    string    `json:"username" validate:"required,alphanum,max=100"`
	Email       string    `json:"email" validate:"required,email,max=100"`
	Fullname    string    `json:"fullname" validate:"required,max=255"`
	Gender      string    `json:"gender" validate:"oneof=male female other"`
	DateOfBirth Date      `json:"date_of_birth" validate:"daterequired"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by"`
}

type ICustomerModel interface {
	Insert(ctx context.Context, username, email, fullname, gender string, dob time.Time) (uuid.UUID, error)
	SelectAll(ctx context.Context) ([]Customer, error)
	SelectById(ctx context.Context, id uuid.UUID) (Customer, error)
	SelectNext(ctx context.Context, customer Customer) ([]Customer, error)
	SelectPrev(ctx context.Context, customer Customer) ([]Customer, error)
	Update(ctx context.Context, payload Customer) (Customer, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type customerModel struct {
	database_connection *sql.DB
	table               string
	fields              string
}

func NewCustomer(db *sql.DB, table_name string) ICustomerModel {
	return &customerModel{
		table:               table_name,
		database_connection: db,
		fields:              "id_text, fullname, gender, email, username, date_of_birth, created_at, created_by",
	}
}

func (model *customerModel) Insert(ctx context.Context, username, email, fullname, gender string, dob time.Time) (uuid.UUID, error) {
	var id uuid.UUID

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return id, err
	}

	id = uuid.New()
	now := time.Now().In(loc)

	if gender == "" {
		gender = "other"
	}

	sql_query :=
		`INSERT INTO customer(
			id,
			id_text,
			username,
			email,
			fullname,
			gender,
			date_of_birth,
			created_at,
			created_by
		)
	VALUES (
			unhex(replace(?, '-', '')),
			UPPER(?),
			?,
			?,
			?,
			?,
			?,
			?,
			?
		)`

	claims, err := auth.ExtractAuthClaims(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = model.database_connection.ExecContext(ctx, sql_query, id, id.String(), username, email, fullname, gender, dob, now, claims.Email)
	if err != nil {
		id = uuid.Nil
	}

	return id, err
}

func (model *customerModel) SelectAll(ctx context.Context) ([]Customer, error) {
	var customers []Customer
	var customer Customer

	claims, err := auth.ExtractAuthClaims(ctx)
	if err != nil {
		return nil, err
	}

	sql_query := fmt.Sprintf("SELECT %s FROM portfolio.customer a WHERE a.created_by=? ORDER BY fullname ASC LIMIT ?", model.fields)
	rows, err := model.database_connection.QueryContext(ctx, sql_query, claims.Email, Max_limit+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		var fullname sql.NullString
		var gender sql.NullString
		var email sql.NullString
		var username sql.NullString
		var date_of_birth sql.NullTime
		var created_at time.Time
		var created_by string

		err = rows.Scan(&id, &fullname, &gender, &email, &username, &date_of_birth, &created_at, &created_by)
		if err != nil {
			return nil, err
		}

		if id.Valid {
			customer.Id, err = uuid.Parse(id.String)
			if err != nil {
				return nil, err
			}
		}

		if fullname.Valid {
			customer.Fullname = fullname.String
		}

		if gender.Valid {
			customer.Gender = gender.String
		}

		if email.Valid {
			customer.Email = email.String
		}

		if username.Valid {
			customer.Username = username.String
		}

		if date_of_birth.Valid {
			customer.DateOfBirth = Date(date_of_birth.Time)
		}

		customer.CreatedAt = created_at
		customer.CreatedBy = created_by

		customers = append(customers, customer)
	}

	return customers, nil
}

func (model *customerModel) SelectById(ctx context.Context, id uuid.UUID) (Customer, error) {
	var customer Customer

	sql_query := fmt.Sprintf("SELECT %s FROM portfolio.customer a WHERE a.id_text=?", model.fields)
	rows, err := model.database_connection.QueryContext(ctx, sql_query, id)
	if err != nil {
		return customer, err
	}
	defer rows.Close()

	if rows.Next() {
		var id sql.NullString
		var fullname sql.NullString
		var gender sql.NullString
		var email sql.NullString
		var username sql.NullString
		var date_of_birth sql.NullTime
		var created_at time.Time
		var created_by string

		err := rows.Scan(&id, &fullname, &gender, &email, &username, &date_of_birth, &created_at, &created_by)
		if err != nil {
			return customer, err
		}

		customer.CreatedAt = created_at
		customer.CreatedBy = created_by

		if id.Valid {
			customer.Id, err = uuid.Parse(id.String)
			if err != nil {
				return customer, err
			}
		}

		if fullname.Valid {
			customer.Fullname = fullname.String
		}

		if gender.Valid {
			customer.Gender = gender.String
		}

		if email.Valid {
			customer.Email = email.String
		}

		if username.Valid {
			customer.Username = username.String
		}

		if date_of_birth.Valid {
			customer.DateOfBirth = Date(date_of_birth.Time)
		}
	}

	return customer, nil
}

func (model *customerModel) SelectNext(ctx context.Context, customer Customer) ([]Customer, error) {
	var customers []Customer

	sql_query := fmt.Sprintf("SELECT %s FROM customer WHERE fullname > ? ORDER BY fullname ASC LIMIT ?", model.fields)
	rows, err := model.database_connection.QueryContext(ctx, sql_query, customer.Fullname, Max_limit+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		var fullname sql.NullString
		var gender sql.NullString
		var email sql.NullString
		var username sql.NullString
		var date_of_birth sql.NullTime
		var created_at time.Time
		var created_by string

		err = rows.Scan(&id, &fullname, &gender, &email, &username, &date_of_birth, &created_at, &created_by)
		if err != nil {
			return nil, err
		}

		customer.CreatedAt = created_at
		customer.CreatedBy = created_by

		if id.Valid {
			customer.Id, err = uuid.Parse(id.String)
			if err != nil {
				return nil, err
			}
		}

		if fullname.Valid {
			customer.Fullname = fullname.String
		}

		if gender.Valid {
			customer.Gender = gender.String
		}

		if email.Valid {
			customer.Email = email.String
		}

		if username.Valid {
			customer.Username = username.String
		}

		if date_of_birth.Valid {
			customer.DateOfBirth = Date(date_of_birth.Time)
		}

		customers = append(customers, customer)
	}

	return customers, nil
}

func (model *customerModel) SelectPrev(ctx context.Context, customer Customer) ([]Customer, error) {
	var customers []Customer

	sql_query := fmt.Sprintf(`
	SELECT b.* FROM (
		SELECT %s FROM portfolio.customer a WHERE a.fullname < ? ORDER BY a.fullname DESC LIMIT ?
		) b
		ORDER BY b.fullname ASC;`, model.fields)
	rows, err := model.database_connection.QueryContext(ctx, sql_query, customer.Fullname, Max_limit+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		var fullname sql.NullString
		var gender sql.NullString
		var email sql.NullString
		var username sql.NullString
		var date_of_birth sql.NullTime
		var created_at time.Time
		var created_by string

		err = rows.Scan(&id, &fullname, &gender, &email, &username, &date_of_birth, &created_at, &created_by)
		if err != nil {
			return nil, err
		}

		customer.CreatedAt = created_at
		customer.CreatedBy = created_by

		if id.Valid {
			customer.Id, err = uuid.Parse(id.String)
			if err != nil {
				return nil, err
			}
		}

		if fullname.Valid {
			customer.Fullname = fullname.String
		}

		if gender.Valid {
			customer.Gender = gender.String
		}

		if email.Valid {
			customer.Email = email.String
		}

		if username.Valid {
			customer.Username = username.String
		}

		if date_of_birth.Valid {
			customer.DateOfBirth = Date(date_of_birth.Time)
		}

		customers = append(customers, customer)
	}

	return customers, nil
}

func (model *customerModel) Update(ctx context.Context, payload Customer) (Customer, error) {
	var updated_customer Customer
	claims, err := auth.ExtractAuthClaims(ctx)
	if err != nil {
		return updated_customer, err
	}

	fields := []string{}
	struct_fields := []interface{}{}

	if payload.Username != "" {
		fields = append(fields, "username=?")
		struct_fields = append(struct_fields, payload.Username)
	}

	if payload.Fullname != "" {
		fields = append(fields, "fullname=?")
		struct_fields = append(struct_fields, payload.Fullname)
	}

	if payload.Email != "" {
		fields = append(fields, "email=?")
		struct_fields = append(struct_fields, payload.Email)
	}

	if payload.Gender != "" {
		fields = append(fields, "gender=?")
		struct_fields = append(struct_fields, payload.Gender)
	}

	empty_date := Date{}
	if payload.DateOfBirth != empty_date {
		fields = append(fields, "date_of_birth=?")
		struct_fields = append(struct_fields, payload.DateOfBirth.Format())
	}

	struct_fields = append(struct_fields, payload.Id.String())
	struct_fields = append(struct_fields, claims.Email)

	tx, err := model.database_connection.BeginTx(ctx, nil)
	if err != nil {
		return updated_customer, err
	}
	defer tx.Rollback()

	sql_query := fmt.Sprintf("UPDATE portfolio.%s SET %s WHERE id_text=? AND created_by=?", model.table, strings.Join(fields, ", "))
	_, err = tx.ExecContext(ctx, sql_query, struct_fields...)
	if err != nil {
		return updated_customer, err
	}

	sql_query = fmt.Sprintf("SELECT %s FROM portfolio.%s WHERE id_text=? AND created_by=?", model.fields, model.table)
	row := tx.QueryRowContext(ctx, sql_query, payload.Id.String(), claims.Email)

	var id sql.NullString
	var fullname sql.NullString
	var gender sql.NullString
	var email sql.NullString
	var username sql.NullString
	var date_of_birth sql.NullTime
	var created_at time.Time
	var created_by string

	err = row.Scan(&id, &fullname, &gender, &email, &username, &date_of_birth, &created_at, &created_by)
	if err != nil {
		return updated_customer, err
	}

	if id.Valid {
		updated_customer.Id, err = uuid.Parse(id.String)
		if err != nil {
			return updated_customer, err
		}
	}

	if fullname.Valid {
		updated_customer.Fullname = fullname.String
	}

	if gender.Valid {
		updated_customer.Gender = gender.String
	}

	if email.Valid {
		updated_customer.Email = email.String
	}

	if username.Valid {
		updated_customer.Username = username.String
	}

	if date_of_birth.Valid {
		updated_customer.DateOfBirth = Date(date_of_birth.Time)
	}

	updated_customer.CreatedAt = created_at
	updated_customer.CreatedBy = created_by

	err = tx.Commit()
	if err != nil {
		return updated_customer, err

	}

	return updated_customer, nil
}

func (model *customerModel) Delete(ctx context.Context, id uuid.UUID) error {
	claims, err := auth.ExtractAuthClaims(ctx)
	if err != nil {
		return err
	}

	sql_query := fmt.Sprintf("DELETE FROM portfolio.%s a WHERE a.id_text=? AND a.created_by=?", model.table)
	_, err = model.database_connection.ExecContext(ctx, sql_query, id, claims.Email)
	if err != nil {
		return err
	}

	return nil
}
