package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mmiftahrzki/go-rest-api/response"
)

type middleware struct {
	handlers []Handler
	router   httprouter.Router
}

type Handler func(w http.ResponseWriter, r *http.Request) error

func New(r httprouter.Router) *middleware {
	return &middleware{
		handlers: []Handler{},
		router:   r,
	}
}

func (m *middleware) Add(h Handler) {
	m.handlers = append(m.handlers, h)
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, v := range m.handlers {
		err := v(w, r)
		if err != nil {
			res := response.New()
			res.Message = err.Error()

			w.Header().Set("Content-Type", "application/json")
			w.Write(res.ToJson())

			return
		}
	}

	m.router.ServeHTTP(w, r)
}
