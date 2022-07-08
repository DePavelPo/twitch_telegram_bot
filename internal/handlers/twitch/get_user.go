package twitch_handler

import (
	"errors"
	"net/http"
	"strings"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

func (twh *TwitchHandler) GetUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	reqDTO := models.GetUserInfoReq{}
	if err := jsoniter.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		logrus.Errorf("failed decode request, error: %v", err)
		WriteErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	authStr := r.Header.Get("Authorization")
	if !strings.HasPrefix(authStr, "Bearer ") {
		err := errors.New("GetUser: token missed")
		logrus.Errorf(err.Error())
		WriteErrorResponse(w, r, http.StatusUnauthorized, err.Error())
		return
	}

	token := authStr[len("Bearer "):]

	res, err := twh.twitchService.GetUser(ctx, token, reqDTO.ID)
	if err != nil {
		logrus.Error(err)
		WriteErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccessData(w, r, res)
}
