package validation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	pkg_validator "github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/middleware"
	"github.com/mmiftahrzki/go-rest-api/model"
)

type jwtContextKey int

const key jwtContextKey = iota

var validator *pkg_validator.Validate

func init() {
	validator = pkg_validator.New()

	validator.RegisterValidation("daterequired", func(fl pkg_validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(model.Date)
		if !ok {
			return false
		}

		if value == model.Date(time.Time{}) {
			return false
		}

		return true
	})
}

func New() middleware.Middleware {
	return validationHandler
}

func validationHandler(next httprouter.Handle) httprouter.Handle {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		r_body, err := io.ReadAll(request.Body)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		customer := &model.Customer{}
		err = json.Unmarshal(r_body, customer)
		if err != nil {
			log.Println(err)

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		err = validator.Struct(customer)
		if err != nil {
			log.Println(err)

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		request = request.WithContext(context.WithValue(request.Context(), key, customer))

		next(writer, request, params)
	}
}

func ExtractCustomerFromContext(ctx context.Context) (*model.Customer, error) {
	customer_value := ctx.Value(key)
	customer, ok := customer_value.(*model.Customer)
	if !ok {
		return nil, errors.New("validation: invalid customer")
	}

	return customer, nil
}

func ExtractUserFromContext(ctx context.Context) (*model.User, error) {
	user_value := ctx.Value(key)
	user, ok := user_value.(*model.User)
	if !ok {
		return nil, errors.New("validation: invalid user")
	}

	return user, nil
}
