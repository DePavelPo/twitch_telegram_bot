package telegram

import (
	"net/http"

	"twitch_telegram_bot/internal/middleware"

	"github.com/sirupsen/logrus"
)

type Response struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func (h *TelegramHandler) GetBotData(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	res, err := h.telegramService.GetBotData(ctx)
	if err != nil {
		logrus.Error(err)
		middleware.WriteErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.WriteSuccessData(w, r, res)
}
