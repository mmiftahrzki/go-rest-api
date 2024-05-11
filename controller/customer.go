package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/mmiftahrzki/go-rest-api/middleware/validation"
	"github.com/mmiftahrzki/go-rest-api/model"
	"github.com/mmiftahrzki/go-rest-api/response"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type ICustomer interface {
	Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	ReadAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	ReadNext(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	ReadPrev(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	ReadById(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
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
				res.Message = fmt.Sprintf("customer dengan username: %s sudah ada", new_customer.Username)

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

	res.Message = "berhasil membuat customer baru"
	res.Data["id"] = id.String()

	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(res.ToJson()))
}

func (c *customer) ReadAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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
	response.Message = "berhasil mendapatkan data customer"

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(response.ToJson())
}

func (c *customer) ReadById(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "id tidak valid"

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

	empty_customer := model.Customer{}
	if customer == empty_customer {
		res.Message = fmt.Sprintf("customer dengan id: %s tidak ditemukan", id)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write(res.ToJson())

		return
	}

	res.Message = "berhasil mendapatkan data customer"
	res.Data["customer"] = customer

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(res.ToJson())
}

func (c *customer) ReadNext(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "id tidak valid"

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
	res.Message = "berhasil mendapatkan data customer"

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res.ToJson())
}

func (c *customer) ReadPrev(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "id tidak valid"

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
	res.Message = "berhasil mendapatkan data customer"

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res.ToJson())
}

func (c *customer) UpdateById(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()
	var err error

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "id tidak valid"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
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

	payload.Id = id

	customer, err := c.model.Update(request.Context(), payload)
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

	res.Message = "berhasil memperbarui data customer"
	res.Data["customer"] = customer

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(res.ToJson())
}

func (c *customer) Delete(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	res := response.New()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "id tidak valid"

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(res.ToJson())

		return
	}

	err = c.model.Delete(request.Context(), id)
	if err != nil {
		log.Println(err)

		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNoContent)
}
