package twitch_handler

import (
	"net/http"
	"twitch_telegram_bot/internal/middleware"

	"github.com/sirupsen/logrus"
)

func (twh *TwitchHandler) GetOAuthToken(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	res, err := twh.twitchService.GetOAuthToken(ctx)
	if err != nil {
		logrus.Error(err)
		middleware.WriteErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.WriteSuccessData(w, r, res)
}
