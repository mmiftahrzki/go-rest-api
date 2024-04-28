package response

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func New() *response {
	return &response{
		Message: "Terjadi kesalahan di sisi penyedia layanan.",
		Data:    map[string]interface{}{},
	}
}

func (r *response) Send(w http.ResponseWriter) {
	w.Write([]byte("sukses"))
}

func (r *response) ToJson() []byte {
	json_enc_r, _ := json.Marshal(r)

	return json_enc_r
}
