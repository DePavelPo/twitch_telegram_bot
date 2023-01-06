package twitch_handler

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func (twh *TwitchHandler) GetUserToken(w http.ResponseWriter, r *http.Request) {

	logrus.Info(r.URL.Query().Get("code"))

	if r.URL.Query().Get("state") != os.Getenv("TWITCH_STATE") {
		logrus.Error("invalid request state")
		return
	}

	// token := r.URL.Query().Get("code")

}
