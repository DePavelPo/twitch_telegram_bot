package twitch_handler

import (
	"net/http"
	"strings"
	"twitch_telegram_bot/internal/middleware"

	"github.com/sirupsen/logrus"
)

func (twh *TwitchHandler) GetUserToken(w http.ResponseWriter, r *http.Request) {

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")

	if code == "" || state == "" {
		err := "empty code or state"
		logrus.Error(err)
		middleware.WriteErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	ctx := r.Context()

	err := twh.twitchUserAuthservice.CheckUserTokensByState(ctx, code, state, strings.Split(scope, " "))
	if err != nil {
		logrus.Error(err)
		middleware.WriteErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.WriteSuccessMessage(w, r, "request was processed successfully")

}
