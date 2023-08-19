package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mmiftahrzki/go-rest-api/controller"
	"github.com/mmiftahrzki/go-rest-api/database"
	"github.com/mmiftahrzki/go-rest-api/middleware"
	"github.com/mmiftahrzki/go-rest-api/model"

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
	customer_controller := controller.NewCustomerController(customer_model)
	router := httprouter.New()
	// response := model.NewResponse()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// response.Message = "the resource you are looking is not found"

		// r_json_encoded, err := json.Marshal(response)
		// if err != nil {
		// 	log.Println(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	w.Write(r_json_encoded)

		// 	return
		// }

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
	})
	router.GET("/api/customers", customer_controller.FindAll)
	router.POST("/api/customers", customer_controller.Create)
	router.GET("/api/customers/:id/next", customer_controller.FindNext)
	router.GET("/api/customers/:id/prev", customer_controller.FindPrev)
	router.GET("/api/customers/:id", customer_controller.FindById)
	router.DELETE("/api/customers/:id", customer_controller.Delete)
	router.PUT("/api/customers/:id", customer_controller.UpdateById)

	auth := middleware.NewAuth(router)
	server := http.Server{
		Addr:    os.Getenv("BASE_URL") + ":" + os.Getenv("PORT"),
		Handler: auth,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
