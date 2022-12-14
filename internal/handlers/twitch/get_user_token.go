package twitch_handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func (twh *TwitchHandler) GetUserToken(w http.ResponseWriter, r *http.Request) {
	fragment := r.URL.Fragment
	logrus.Info(r.URL.RequestURI())
	logrus.Info(r.URL.RawFragment)
	logrus.Info(fragment)
	// userToken := fragment[strings.LastIndex(fragment, "access_token="):strings.Index(fragment, "&")]
	// logrus.Info(userToken)
}
