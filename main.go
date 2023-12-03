package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mmiftahrzki/go-rest-api/controller"
	"github.com/mmiftahrzki/go-rest-api/database"
	"github.com/mmiftahrzki/go-rest-api/middleware"
	"github.com/mmiftahrzki/go-rest-api/middleware/auth"
	"github.com/mmiftahrzki/go-rest-api/model"
	"github.com/mmiftahrzki/go-rest-api/response"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

const Max_data_per_call = 10

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln(err)
	}

	db, err := database.NewDB()
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	customer_model := model.NewCustomerModel(db)
	// customer_controller := controller.NewCustomerController(customer_model, 10)
	customer_controller := controller.NewCustomer(customer_model)
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := response.New()
		response.Message = "the resource you are looking is not found"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(response.ToJson())
	})
	router.GET("/api/customers", customer_controller.FindAll)
	router.POST("/api/customers", customer_controller.Create)
	router.GET("/api/customers/:id/next", customer_controller.FindNext)
	router.GET("/api/customers/:id/prev", customer_controller.FindPrev)
	router.GET("/api/customers/:id", customer_controller.FindById)
	router.DELETE("/api/customers/:id", customer_controller.Delete)
	router.PUT("/api/customers/:id", customer_controller.UpdateById)

	middleware := middleware.New(*router)
	auth := auth.New()

	middleware.Add(auth)

	server := http.Server{
		Addr: os.Getenv("BASE_URL") + ":" + os.Getenv("PORT"),
		// Handler: auth,
		Handler: middleware,
		// Handler: router,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
