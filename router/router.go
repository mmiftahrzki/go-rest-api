package router

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/response"
)

func New() *httprouter.Router {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// response := response.New()
		res := response.New()
		res.Message = "the resource you are looking is not found"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write(res.ToJson())
	})

	return router
}
