package twitch_handler

import (
	"net/http"
	"twitch_telegram_bot/internal/middleware"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

func (twh *TwitchHandler) GetActiveStreamInfoByUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	reqDTO := models.GetActiveStreamInfoByUserReq{}
	if err := jsoniter.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		logrus.Errorf("failed decode request, error: %v", err)
		middleware.WriteErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// authStr := r.Header.Get("Authorization")
	// if !strings.HasPrefix(authStr, "Bearer ") {
	// 	err := errors.New("GetActiveStreamInfoByUser: token missed")
	// 	logrus.Errorf(err.Error())
	// 	middleware.WriteErrorResponse(w, r, http.StatusUnauthorized, err.Error())
	// 	return
	// }

	res, err := twh.twitchService.GetActiveStreamInfoByUser(ctx, reqDTO.ID)
	if err != nil {
		logrus.Error(err)
		middleware.WriteErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.WriteSuccessData(w, r, res)
}
