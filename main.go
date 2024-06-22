package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/controller"
	"github.com/mmiftahrzki/go-rest-api/database"
	"github.com/mmiftahrzki/go-rest-api/handler"
	"github.com/mmiftahrzki/go-rest-api/middleware"
	auth_pkg "github.com/mmiftahrzki/go-rest-api/middleware/auth"
	"github.com/mmiftahrzki/go-rest-api/middleware/validation"
	"github.com/mmiftahrzki/go-rest-api/model"
	router_pkg "github.com/mmiftahrzki/go-rest-api/router"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln(err)
	}

	db := database.GetDatabaseConnection()
	defer db.Close()

	model_customer := model.NewCustomer(db, "customer")
	controller_customer := controller.NewCustomer(model_customer)
	router := router_pkg.New()
	auth := auth_pkg.New()
	customerValidation := validation.New()

	helloWorld := router_pkg.Endpoint{Path: "/", Method: http.MethodGet}

	signUp := router_pkg.Endpoint{Path: "/api/auth/signup", Method: http.MethodPost}
	signIn := router_pkg.Endpoint{Path: "/api/auth/signin", Method: http.MethodPost}

	getToken := router_pkg.Endpoint{Path: "/api/auth/token", Method: http.MethodPost}

	createCustomer := router_pkg.Endpoint{Path: "/api/customers", Method: http.MethodPost, Middlewares: []middleware.Middleware{auth, customerValidation}}
	getAllCustomers := router_pkg.Endpoint{Path: "/api/customers", Method: http.MethodGet, Middlewares: []middleware.Middleware{auth}}
	getCustomerById := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodGet, Middlewares: []middleware.Middleware{auth}}
	updateCustomer := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodPut, Middlewares: []middleware.Middleware{auth}}
	deleteCustomer := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodDelete, Middlewares: []middleware.Middleware{auth}}

	router.AddRoute(helloWorld, func(writer http.ResponseWriter, request *http.Request, parameters httprouter.Params) {
		writer.Header().Set("Content-Type", "text/html")
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`<!DOCTYPE html><html lang="en"><head><meta charset="UTF-8" /><meta name="viewport" content="width=device-width, initial-scale=1.0" /><title>Greetings!</title></head><body><h1>Hello, Wordl!</h1></body></html>`))
	})
	router.AddRoute(signUp, handler.CreateUser)
	router.AddRoute(signIn, handler.ReadUser)

	router.AddRoute(getToken, auth_pkg.Token)

	router.AddRoute(createCustomer, controller_customer.Create)
	router.AddRoute(getAllCustomers, controller_customer.ReadAll)
	router.AddRoute(router_pkg.Endpoint{Path: "/api/customers/:id/next", Method: http.MethodGet}, controller_customer.ReadNext)
	router.AddRoute(router_pkg.Endpoint{Path: "/api/customers/:id/prev", Method: http.MethodGet}, controller_customer.ReadPrev)
	router.AddRoute(updateCustomer, controller_customer.UpdateById)
	router.AddRoute(deleteCustomer, controller_customer.Delete)
	router.AddRoute(getCustomerById, controller_customer.ReadById)

	server := http.Server{
		Addr:    os.Getenv("BASE_URL") + ":" + os.Getenv("PORT"),
		Handler: router,
	}

	fmt.Println("Listening on:", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
