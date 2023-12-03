package validation

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mmiftahrzki/go-rest-api/middleware"
	"github.com/mmiftahrzki/go-rest-api/model"
)

func New() middleware.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		paths := strings.Split(r.URL.Path, "/")

		if len(paths) > 0 {
			v := validator.New()
			endpoint := paths[2]

			if endpoint == "customers" {
				var customer *model.Customer

				json_decoder := json.NewDecoder(r.Body)
				err := json_decoder.Decode(&customer)
				if err != nil {
					return err
				}

				err = v.Struct(customer)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func ValidateDateType(field reflect.Value) interface{} {
	_, ok := field.Interface().(model.Date)
	if ok {
		return nil
	}

	return nil
}
