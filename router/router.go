package router

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/middleware"
	"github.com/mmiftahrzki/go-rest-api/response"
)

type Endpoint struct {
	Middlewares []middleware.Middleware
	Method      string
	Path        string
}

type Router struct {
	endpoints  map[string]Endpoint
	httprouter *httprouter.Router
}

func New() *Router {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := response.New()
		response.Message = "sumber daya yang Anda cari tidak ditemukan"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write(response.ToJson())
	})
	router.MethodNotAllowed = router.NotFound

	return &Router{
		httprouter: router,
		endpoints:  map[string]Endpoint{},
	}
}

func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var handler_funcs http.HandlerFunc = router.httprouter.ServeHTTP

	key := fmt.Sprintf("%s%s", request.URL.Path, request.Method)
	for i := len(router.endpoints[key].Middlewares) - 1; i >= 0; i-- {
		handler_funcs = router.endpoints[key].Middlewares[i](handler_funcs)
	}

	handler_funcs.ServeHTTP(writer, request)
}

func (router *Router) AddRoute(endpoint Endpoint, handle httprouter.Handle) {
	router.httprouter.Handle(endpoint.Method, endpoint.Path, handle)

	key := fmt.Sprintf("%s%s", endpoint.Path, endpoint.Method)
	router.endpoints[key] = endpoint
}
