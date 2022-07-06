package twitch_client

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
)

type TwitchClient struct {
}

func NewTwitchClient() *TwitchClient {
	return &TwitchClient{}
}

func (twc *TwitchClient) TwitchOAuth(ctx context.Context) (data *models.TwitchOathResponse, err error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("client_id", os.Getenv("TWITCH_CLIENT_ID"))
	query.Add("client_secret", os.Getenv("TWITCH_SECRET"))
	query.Add("grant_type", "client_credentials")
	req.URL.RawQuery = query.Encode()

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	readedResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var tokenInfo models.TwitchOathResponse
	err = jsoniter.Unmarshal(readedResp, &tokenInfo)
	if err != nil {
		return
	}

	data = &tokenInfo

	return
}
