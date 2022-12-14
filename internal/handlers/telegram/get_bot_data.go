package telegram

import (
	"net/http"

	"twitch_telegram_bot/internal/middleware"
)

type Response struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func (h *TelegramHandler) GetBotData(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	res, err := h.telegramService.GetBotCommands(ctx)
	if err != nil {

		middleware.WriteErrorResponse(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	middleware.WriteSuccessData(w, r, res)
}
