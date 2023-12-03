package model

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mmiftahrzki/go-rest-api/middleware/auth"
)

const Max_limit int = 10

// type date time.Time
type Date time.Time

func (j *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
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
	Id          uuid.UUID `json:"id" validate:"required,uuid4"`
	Username    string    `json:"username" validate:"required,alphanum,max=100"`
	Email       string    `json:"email" validate:"required,email,max=100"`
	Fullname    string    `json:"fullname" validate:"required,max=255"`
	Gender      string    `json:"gender" validate:"oneof=male female other"`
	DateOfBirth Date      `json:"date_of_birth" validate:"date"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by" validate:"required,email,max=320"`
}

type ICustomerModel interface {
	Insert(ctx context.Context, username, email, fullname, gender string, dob time.Time) error
	SelectAll(ctx context.Context) ([]Customer, error)
	SelectById(ctx context.Context, id uuid.UUID) (Customer, error)
	SelectNext(ctx context.Context, customer Customer) ([]Customer, error)
	SelectPrev(ctx context.Context, customer Customer) ([]Customer, error)
	Update(ctx context.Context, customer, payload Customer) (Customer, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type customerModel struct {
	database_connection *sql.DB
	fields              string
}

func NewCustomerModel(db *sql.DB) ICustomerModel {
	return &customerModel{
		database_connection: db,
		fields:              "id_text, fullname, gender, email, username, date_of_birth, created_at, created_by",
	}
}

func (model *customerModel) Insert(ctx context.Context, username, email, fullname, gender string, dob time.Time) error {
	id := uuid.New()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return err
	}

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

	// _, err = model.database_connection.ExecContext(ctx, sql_query, id, id.String(), username, email, fullname, gender, dob, now, middleware.Claims.Email)
	_, err = model.database_connection.ExecContext(ctx, sql_query, id, id.String(), username, email, fullname, gender, dob, now, auth.Claims.Email)

	return err
}

func (model *customerModel) SelectAll(ctx context.Context) ([]Customer, error) {
	var customers []Customer

	sql_query := fmt.Sprintf("SELECT %s FROM portfolio.customer a WHERE a.created_by=? ORDER BY fullname ASC LIMIT ?", model.fields)
	// rows, err := model.database_connection.QueryContext(ctx, sql_query, middleware.Claims.Email, Max_limit+1)
	rows, err := model.database_connection.QueryContext(ctx, sql_query, auth.Claims.Email, Max_limit+1)
	if err != nil {
		return nil, err
	}
	defer func() {
		rows.Close()
	}()

	var customer Customer

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
	defer func() {
		fmt.Println("rows closed")
		rows.Close()
	}()

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

func (model *customerModel) Update(ctx context.Context, customer, payload Customer) (Customer, error) {
	var updated_customer Customer
	fields := []string{}
	struct_fields := []interface{}{customer.Id}

	if !reflect.ValueOf(payload.Fullname).IsZero() {
		fields = append(fields, "fullname=?")
		struct_fields = append(struct_fields, payload.Fullname)
	}

	if !reflect.ValueOf(payload.Gender).IsZero() {
		fields = append(fields, "gender=?")
		struct_fields = append(struct_fields, payload.Gender)
	}

	if !reflect.ValueOf(payload.Email).IsZero() {
		fields = append(fields, "email=?")
		struct_fields = append(struct_fields, payload.Email)
	}

	if !reflect.ValueOf(payload.Username).IsZero() {
		fields = append(fields, "username=?")
		struct_fields = append(struct_fields, payload.Username)
	}

	if !reflect.ValueOf(payload.DateOfBirth).IsZero() {
		fields = append(fields, "date_of_birth=?")
		struct_fields = append(struct_fields, payload.DateOfBirth.Format())
	}

	sql_query := fmt.Sprintf("UPDATE customer SET %s WHERE id_text=?", strings.Join(fields, ", "))
	_, err := model.database_connection.ExecContext(ctx, sql_query, struct_fields...)
	if err != nil {
		return updated_customer, err
	}

	updated_customer, err = model.SelectById(ctx, customer.Id)
	if err != nil {
		return updated_customer, err
	}

	return updated_customer, nil
}

func (model *customerModel) Delete(ctx context.Context, id uuid.UUID) error {
	sql_query := "DELETE FROM portfolio.customer a WHERE a.id_text=? AND a.created_by=?"
	// _, err := model.database_connection.ExecContext(ctx, sql_query, id, middleware.Claims.Email)
	_, err := model.database_connection.ExecContext(ctx, sql_query, id, auth.Claims.Email)
	if err != nil {
		return err
	}

	return nil
}
