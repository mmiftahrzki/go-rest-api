package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

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

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&new_customer)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	err = c.model.Insert(req.Context(), new_customer.Username, new_customer.Email, new_customer.Fullname, new_customer.Gender, new_customer.DateOfBirth)
	if err != nil {
		mysql_error, ok := err.(*mysql.MySQLError)
		if ok {
			res := model.NewResponse()

			if mysql_error.Number == 1062 {
				res.Message = fmt.Sprintf("customer with username: %s already exist", new_customer.Username)

				res_json, err := json.Marshal(res)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write(res_json)

				return
			}
		}

		log.Println(err)

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

	res_json, err := json.Marshal(res)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res_json)
}

func (c *customerController) FindById(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res_json)

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

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res_json)

		return
	}

	res.Message = "success retrieve customer data"
	res.Data["customer"] = customer

	res_json, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res_json)
}

func (c *customerController) FindNext(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res_json)

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

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res_json)

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

	res_json, err := json.Marshal(res)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res_json)
}

func (c *customerController) FindPrev(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res_json)

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

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res_json)

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

	res_json, err := json.Marshal(res)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res_json)
}

func (c *customerController) UpdateById(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	_, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res_json)

		return
	}

	payload := model.Customer{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&payload)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	if reflect.ValueOf(payload).IsZero() {
		res.Data["customer"] = payload

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res_json)

		return
	}

	// customer, err := c.model.SelectById(req.Context(), id)
	// if err != nil {
	// 	log.Println(err)

	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

	// 	return
	// }

	// if reflect.ValueOf(customer).IsZero() {
	// 	res.Message = fmt.Sprintf("customer with id: %s not found", id)

	// 	res_json, err := json.Marshal(res)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	// 	}

	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusNotFound)
	// 	w.Write(res_json)

	// 	return
	// }

	// customer, err = c.model.UpdateById(req.Context(), customer)
}

func (c *customerController) Delete(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res := model.NewResponse()

	id, err := uuid.Parse(params.ByName("id"))
	if err != nil {
		log.Println(err)

		res.Message = "invalid id"

		res_json, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res_json)

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

		res_json, err := json.Marshal(res)
		if err != nil {
			log.Println(err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res_json)

		return
	}

	if customer.CreatedBy != middleware.Claims.Email {
		res.Message = http.StatusText(http.StatusUnauthorized)

		res_json, err := json.Marshal(res)
		if err != nil {
			log.Println(err)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(res_json)

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

	res_json, err := json.Marshal(res)
	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res_json)
}
