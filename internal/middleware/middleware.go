package middleware

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

type Response struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

type MessageResponse struct {
	Message string `json:"msg"`
}

func WriteSuccessData(w http.ResponseWriter, r *http.Request, data interface{}) {
	_ = jsoniter.NewEncoder(w).Encode(Response{
		Data: data,
	})

	w.WriteHeader(200)
}

func WriteSuccessMessage(w http.ResponseWriter, r *http.Request, data string) {
	_ = jsoniter.NewEncoder(w).Encode(Response{
		Data: MessageResponse{
			Message: data,
		},
	})

	w.WriteHeader(200)
}

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, errCode int, err string) {
	_ = jsoniter.NewEncoder(w).Encode(Response{
		Error: err,
	})

	w.WriteHeader(errCode)
}
