package router

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/response"
)

type Endpoint struct {
	middlewares []Middleware
	Method      string
	Path        string
}

type Myrouter struct {
	endpoints  map[string]Endpoint
	httprouter *httprouter.Router
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func New() *Myrouter {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := response.New()
		response.Message = "the resource you are looking is not found"

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write(response.ToJson())
	})

	return &Myrouter{
		httprouter: router,
		endpoints:  map[string]Endpoint{},
	}
}

func (mr *Myrouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	content_type := request.Header.Get("Content-Type")
	if content_type != "application/json" {
		writer.WriteHeader(http.StatusBadRequest)

		return
	}

	var handler_funcs http.HandlerFunc = mr.httprouter.ServeHTTP
	h, _, _ := mr.httprouter.Lookup(request.Method, request.URL.Path)
	key := fmt.Sprintf("%p", h)
	for i := len(mr.endpoints[key].middlewares) - 1; i >= 0; i-- {
		handler_funcs = mr.endpoints[key].middlewares[i](handler_funcs)
	}

	handler_funcs.ServeHTTP(writer, request)
}

func (mr *Myrouter) AddRoute(endpoint Endpoint, handle httprouter.Handle) {
	mr.httprouter.Handle(endpoint.Method, endpoint.Path, handle)

	h, _, _ := mr.httprouter.Lookup(endpoint.Method, endpoint.Path)
	key := fmt.Sprintf("%p", h)
	mr.endpoints[key] = endpoint
}

func (mr *Myrouter) AddMiddleware(m Middleware) {
	if len(mr.endpoints) == 0 {
		panic("router: tidak bisa menyisipkan middleware karena belum ada endpoint yang dirutekan")
	}

	for k := range mr.endpoints {
		endpoint := mr.endpoints[k]
		endpoint.middlewares = append(endpoint.middlewares, m)
		mr.endpoints[k] = endpoint
	}
}

func (mr *Myrouter) AddMiddlewareExcept(m Middleware, endpoint Endpoint) {
	if len(mr.endpoints) == 0 {
		panic("router: tidak bisa menyisipkan middleware karena belum ada endpoint yang dirutekan")
	}

	for k := range mr.endpoints {
		h, _, _ := mr.httprouter.Lookup(endpoint.Method, endpoint.Path)
		key := fmt.Sprintf("%p", h)
		if key == k {
			continue
		}
		
		endpoint := mr.endpoints[k]
		endpoint.middlewares = append(endpoint.middlewares, m)
		mr.endpoints[k] = endpoint
	}
}

func (mr *Myrouter) AddMiddlewareOnly(m Middleware, endpoint Endpoint) {
	if len(mr.endpoints) == 0 {
		panic("router: tidak bisa menyisipkan middleware karena belum ada endpoint yang dirutekan")
	}
	
	h, _, _ := mr.httprouter.Lookup(endpoint.Method, endpoint.Path)
	key := fmt.Sprintf("%p", h)
	endpoint = mr.endpoints[key]
	endpoint.middlewares = append(endpoint.middlewares, m)
	mr.endpoints[key] = endpoint
}
