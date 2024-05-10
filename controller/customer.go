package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/mmiftahrzki/go-rest-api/middleware/auth"
	"github.com/mmiftahrzki/go-rest-api/middleware/validation"
	"github.com/mmiftahrzki/go-rest-api/model"
	"github.com/mmiftahrzki/go-rest-api/response"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type ICustomer interface {
	Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindNext(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindPrev(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindById(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	UpdateById(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
}

type customer struct {
	model model.ICustomerModel
}

func NewCustomer(model model.ICustomerModel) ICustomer {
	return &customer{
		model: model,
	}
}

func (c *customer) Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()
	new_customer, err := validation.ExtractCustomerFromContext(request.Context())
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	id, err := c.model.Insert(request.Context(), new_customer.Username, new_customer.Email, new_customer.Fullname, new_customer.Gender, time.Time(new_customer.DateOfBirth))
	if err != nil {
		log.Println(err)

		mysql_error, ok := err.(*mysql.MySQLError)
		if ok {
			if mysql_error.Number == 1062 {
				res.Message = fmt.Sprintf("customer with username: %s already exist", new_customer.Username)

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

	res.Message = "customer created successfully"
	res.Data["id"] = id.String()

	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(res.ToJson()))
}

func (c *customer) FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	response := response.New()

	customers, err := c.model.SelectAll(request.Context())
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(response.ToJson()))

		return
	}

	if len(customers) == model.Max_limit+1 {
		response.Data["__next"] = fmt.Sprintf("%s:%s/api/customers/%s/next", os.Getenv("BASE_URL"), os.Getenv("PORT"), customers[model.Max_limit-1].Id)

		customers = customers[:model.Max_limit]
	}

	response.Data["customers"] = customers
	response.Message = "success retrieving customers data"

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(response.ToJson())
}

func (c *customer) FindById(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(request.Context(), id)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = fmt.Sprintf("customer with id: %s is not found", id)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write(res.ToJson())

		return
	}

	res.Message = "success retrieve customer data"
	res.Data["customer"] = customer

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(res.ToJson())
}

func (c *customer) FindNext(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(request.Context(), id)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = http.StatusText(http.StatusNotFound)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write(res.ToJson())

		return
	}

	customers, err := c.model.SelectNext(request.Context(), customer)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if len(customers) == model.Max_limit+1 {
		res.Data["__prev"] = fmt.Sprintf("%s:%s/api/customers/%s/prev", os.Getenv("BASE_URL"), os.Getenv("PORT"), (customers[0].Id).String())
		res.Data["__next"] = fmt.Sprintf("%s:%s/api/customers/%s/next", os.Getenv("BASE_URL"), os.Getenv("PORT"), (customers[model.Max_limit-1].Id).String())

		customers = customers[:model.Max_limit]
	}

	res.Data["customers"] = customers
	res.Message = "success retrieving customers data"

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res.ToJson())
}

func (c *customer) FindPrev(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(request.Context(), id)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = http.StatusText(http.StatusNotFound)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write(res.ToJson())

		return
	}

	customers, err := c.model.SelectPrev(request.Context(), customer)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if len(customers) == model.Max_limit+1 {
		res.Data["__prev"] = fmt.Sprintf("%s:%s/api/customers/%s/prev", os.Getenv("BASE_URL"), os.Getenv("PORT"), (customers[1].Id).String())

		customers = customers[1 : model.Max_limit+1]
	}

	if len(customers) > 0 {
		res.Data["__next"] = fmt.Sprintf("%s:%s/api/customers/%s/next", os.Getenv("BASE_URL"), os.Getenv("PORT"), (customers[len(customers)-1].Id).String())
	}

	res.Data["customers"] = customers
	res.Message = "success retrieving customers data"

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res.ToJson())
}

func (c *customer) UpdateById(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(request.Context(), id)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = fmt.Sprintf("customer with id: %s not found", id)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write(res.ToJson())

		return
	}

	payload := model.Customer{}
	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		log.Println(err)

		res.Message = http.StatusText(http.StatusBadRequest)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	if reflect.ValueOf(payload).IsZero() {
		res.Message = http.StatusText(http.StatusBadRequest)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	// if customer.CreatedBy != auth.Claims.Email {
	request.Context().Value("")
	if customer.CreatedBy != "" {
		res.Message = "you can't modify someone else's resource"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write(res.ToJson())

		return
	}

	customer, err = c.model.Update(request.Context(), customer, payload)
	if err != nil {
		log.Println(err)

		mysql_error, ok := err.(*mysql.MySQLError)
		if ok {
			if mysql_error.Number == 1292 {
				res.Message = http.StatusText(http.StatusBadRequest)

				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write(res.ToJson())
			}

			return
		}

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	res.Message = "success updating customer data"
	res.Data["customer"] = customer

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res.ToJson())
}

func (c *customer) Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	claims, err := auth.ExtractAuthClaims(request.Context())
	if err != nil {
		log.Println(err)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	res := response.New()
	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(request.Context(), id)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = "Customer deleted successfully!"

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(res.ToJson())

		return
	}

	// var claims auth.JwtClaims
	// bearer := request.Context().Value(auth.JWTContextKey)
	// claims, _ = bearer.(auth.JwtClaims)

	if customer.CreatedBy != claims.Email {
		res.Message = "you can't modify someone else's resource"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write(res.ToJson())

		return
	}

	// err = c.model.Delete(request.Context(), customer.Id)
	// if err != nil {
	// 	log.Println(err)

	// 	writer.WriteHeader(http.StatusInternalServerError)
	// 	writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

	// 	return
	// }

	res.Message = "Customer deleted successfully!"

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(res.ToJson())
}
