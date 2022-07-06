package twitch_handler

import (
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func (twh *TwitchHandler) GetOAuthToken(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	res, err := twh.twitchService.GetOAuthToken(ctx)
	if err != nil {
		logrus.Error(err)
		WriteErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccessData(w, r, res)
}

// TODO: moved to another package
func WriteSuccessData(w http.ResponseWriter, r *http.Request, data interface{}) {
	_ = jsoniter.NewEncoder(w).Encode(Response{
		Data: data,
	})

	w.WriteHeader(200)
}

// TODO: moved to another package
func WriteErrorResponse(w http.ResponseWriter, r *http.Request, errCode int, err string) {
	_ = jsoniter.NewEncoder(w).Encode(Response{
		Error: err,
	})

	w.WriteHeader(errCode)
}
