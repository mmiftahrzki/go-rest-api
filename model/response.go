package model

import (
	"net/http"
)

type response struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func NewResponse() response {
	return response{
		Message: "Terjadi kesalahan di sisi penyedia layanan.",
		Data:    map[string]interface{}{},
	}
}

func (r *response) Send(w http.ResponseWriter) {
	w.Write([]byte("sukses"))
}

// func (r *response) toJson() ([]byte, error) {
// 	data_encoded, err := json.Marshal(r)
// 	if err != nil {
// 		log.Println(err)
// 		w.Header().Set("Content-Type", "plain/text")
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
// 	}

// 	return data_encoded, err
// }

// func (r *response) WriteResponse(w http.ResponseWriter) {
// 	r_json, err := json.Marshal(r)
// 	if err != nil {
// 		panic(err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	fmt.Fprint(w, string(r_json))
// }
