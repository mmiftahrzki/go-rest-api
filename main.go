package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/mmiftahrzki/go-rest-api/controller"
	"github.com/mmiftahrzki/go-rest-api/database"
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

	db := database.New()
	defer db.Close()
	model_customer := model.NewCustomer(db, "customer_test")
	controller_customer := controller.NewCustomer(model_customer)

	createCustomer := router_pkg.Endpoint{Path: "/api/customers", Method: http.MethodPost}
	getAllCustomers := router_pkg.Endpoint{Path: "/api/customers", Method: http.MethodGet}
	updateCustomer := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodPut}
	deleteCustomer := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodDelete}
	getToken := router_pkg.Endpoint{Path: "/api/auth/token", Method: http.MethodPost}
	getCustomerById := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodGet}

	router := router_pkg.New()
	auth := auth_pkg.New()
	customerValidation := validation.New()

	router.AddRoute(createCustomer, controller_customer.Create)
	router.AddRoute(getAllCustomers, controller_customer.FindAll)
	router.AddRoute(router_pkg.Endpoint{Path: "/api/customers/:id/next", Method: http.MethodGet}, controller_customer.FindNext)
	router.AddRoute(router_pkg.Endpoint{Path: "/api/customers/:id/prev", Method: http.MethodGet}, controller_customer.FindPrev)
	router.AddRoute(updateCustomer, controller_customer.UpdateById)
	router.AddRoute(deleteCustomer, controller_customer.Delete)
	router.AddRoute(getCustomerById, controller_customer.FindById)
	router.AddRoute(getToken, auth_pkg.Token)

	router.AddMiddlewareExcept(auth, getToken)
	router.AddMiddlewareOnly(customerValidation, createCustomer)

	server := http.Server{
		Addr:    os.Getenv("BASE_URL") + ":" + os.Getenv("PORT"),
		Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
