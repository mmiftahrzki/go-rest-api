package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/mmiftahrzki/go-rest-api/middleware"
	"github.com/mmiftahrzki/go-rest-api/model"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type ICustomerController interface {
	Create(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindAll(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindNext(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindPrev(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	FindById(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	UpdateById(writer http.ResponseWriter, request *http.Request, params httprouter.Params)
	Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params)
}

type customerController struct {
	model model.ICustomerModel
}

func NewCustomerController(model model.ICustomerModel) ICustomerController {
	return &customerController{
		model: model,
	}
}

func (c *customerController) Create(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var new_customer model.Customer
	res := model.NewResponse()

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&new_customer)
	if err != nil {
		log.Println(err)

		res.Message = http.StatusText(http.StatusBadRequest)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	err = c.model.Insert(req.Context(), new_customer.Username, new_customer.Email, new_customer.Fullname, new_customer.Gender, time.Time(new_customer.DateOfBirth))
	if err != nil {
		log.Println(err)

		mysql_error, ok := err.(*mysql.MySQLError)
		if ok {
			if mysql_error.Number == 1062 {
				res.Message = fmt.Sprintf("customer with username: %s already exist", new_customer.Username)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write(res.ToJson())

				return
			}
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *customerController) FindAll(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	customers, err := c.model.SelectAll(req.Context())
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	res := model.NewResponse()

	if len(customers) == model.Max_limit+1 {
		res.Data["__next"] = fmt.Sprintf("%s:%s/api/customers/%s/next", os.Getenv("BASE_URL"), os.Getenv("PORT"), customers[model.Max_limit-1].Id)

		customers = customers[:model.Max_limit]
	}

	res.Data["customers"] = customers
	res.Message = "success retrieving customers data"

	w.Header().Set("Content-Type", "application/json")
	w.Write(res.ToJson())
}

func (c *customerController) FindById(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(req.Context(), id)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = fmt.Sprintf("customer with id: %s not found", id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res.ToJson())

		return
	}

	res.Message = "success retrieve customer data"
	res.Data["customer"] = customer

	w.Header().Set("Content-Type", "application/json")
	w.Write(res.ToJson())
}

func (c *customerController) FindNext(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(req.Context(), id)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = http.StatusText(http.StatusNotFound)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res.ToJson())

		return
	}

	customers, err := c.model.SelectNext(req.Context(), customer)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if len(customers) == model.Max_limit+1 {
		res.Data["__prev"] = fmt.Sprintf("%s:%s/api/customers/%s/prev", os.Getenv("BASE_URL"), os.Getenv("PORT"), (customers[0].Id).String())
		res.Data["__next"] = "localhost:3000/api/customers/" + (customers[model.Max_limit-1].Id).String() + "/next"

		customers = customers[:model.Max_limit]
	}

	res.Data["customers"] = customers
	res.Message = "success retrieving customers data"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res.ToJson())
}

func (c *customerController) FindPrev(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(req.Context(), id)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = http.StatusText(http.StatusNotFound)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res.ToJson())

		return
	}

	customers, err := c.model.SelectPrev(req.Context(), customer)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res.ToJson())
}

func (c *customerController) UpdateById(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(req.Context(), id)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = fmt.Sprintf("customer with id: %s not found", id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res.ToJson())

		return
	}

	payload := model.Customer{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		log.Println(err)

		res.Message = http.StatusText(http.StatusBadRequest)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	if reflect.ValueOf(payload).IsZero() {
		res.Message = http.StatusText(http.StatusBadRequest)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	if customer.CreatedBy != middleware.Claims.Email {
		res.Message = "you can't modify someone else's resource"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(res.ToJson())

		return
	}

	customer, err = c.model.Update(req.Context(), customer, payload)
	if err != nil {
		log.Println(err)

		mysql_error, ok := err.(*mysql.MySQLError)
		if ok {
			if mysql_error.Number == 1292 {
				res.Message = http.StatusText(http.StatusBadRequest)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write(res.ToJson())
			}

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	res.Message = "success updating customer data"
	res.Data["customer"] = customer

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res.ToJson())
}

func (c *customerController) Delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.ToJson())

		return
	}

	customer, err := c.model.SelectById(req.Context(), id)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(customer).IsZero() {
		res.Message = "Customer deleted successfully!"

		w.Header().Set("Content-Type", "application/json")
		w.Write(res.ToJson())

		return
	}

	if customer.CreatedBy != middleware.Claims.Email {
		res.Message = "you can't modify someone else's resource"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(res.ToJson())

		return
	}

	err = c.model.Delete(req.Context(), customer.Id)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	res.Message = "Customer deleted successfully!"

	w.Header().Set("Content-Type", "application/json")
	w.Write(res.ToJson())
}
