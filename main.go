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

	db := database.GetDatabaseConnection()
	// db := database.New()
	defer db.Close()

	model_customer := model.NewCustomer(db, "customer")
	controller_customer := controller.NewCustomer(model_customer)

	createUser := router_pkg.Endpoint{Path: "/api/users", Method: http.MethodPost}
	signIn := router_pkg.Endpoint{Path: "/api/auth/signin", Method: http.MethodPost}
	createCustomer := router_pkg.Endpoint{Path: "/api/customers", Method: http.MethodPost}
	getAllCustomers := router_pkg.Endpoint{Path: "/api/customers", Method: http.MethodGet}
	getCustomerById := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodGet}
	updateCustomer := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodPut}
	deleteCustomer := router_pkg.Endpoint{Path: "/api/customers/:id", Method: http.MethodDelete}
	getToken := router_pkg.Endpoint{Path: "/api/auth/token", Method: http.MethodPost}

	router := router_pkg.New()
	auth := auth_pkg.New()
	customerValidation := validation.New()

	router.AddRoute(createUser, controller.CreateUser)
	router.AddRoute(signIn, controller.ReadUser)
	router.AddRoute(createCustomer, controller_customer.Create)
	router.AddRoute(getAllCustomers, controller_customer.ReadAll)
	router.AddRoute(router_pkg.Endpoint{Path: "/api/customers/:id/next", Method: http.MethodGet}, controller_customer.ReadNext)
	router.AddRoute(router_pkg.Endpoint{Path: "/api/customers/:id/prev", Method: http.MethodGet}, controller_customer.ReadPrev)
	router.AddRoute(updateCustomer, controller_customer.UpdateById)
	router.AddRoute(deleteCustomer, controller_customer.Delete)
	router.AddRoute(getCustomerById, controller_customer.ReadById)
	router.AddRoute(getToken, auth_pkg.Token)

	router.InsertMiddlewareExcept(auth, signIn)
	router.InsertMiddlewareOnly(customerValidation, createCustomer)

	server := http.Server{
		Addr:    os.Getenv("BASE_URL") + ":" + os.Getenv("PORT"),
		Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
