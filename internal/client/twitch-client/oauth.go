package twitch_client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const twitchSchemeHost string = "https://id.twitch.tv"

type TwitchClient struct {
}

func NewTwitchClient() *TwitchClient {
	return &TwitchClient{}
}

func (twc *TwitchClient) TwitchOAuthGetToken(ctx context.Context) (data *models.TwitchOatGetTokenhResponse, err error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("POST", twitchSchemeHost+"/oauth2/token", nil)
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

	var tokenInfo models.TwitchOatGetTokenhResponse
	err = jsoniter.Unmarshal(readedResp, &tokenInfo)
	if err != nil {
		return
	}

	data = &tokenInfo

	return
}

func (twc *TwitchClient) TwitchOAuthValidateToken(ctx context.Context, token string) (data *models.TwitchOatValidateTokenhResponse, err error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", twitchSchemeHost+"/oauth2/validate", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	readedResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {

			var unauthorizedResp models.ValidateTokenInvalid
			err = jsoniter.Unmarshal(readedResp, &unauthorizedResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New("token invalid")
		}

		return nil, errors.Errorf("get twitch streams failed with status code: %d", resp.StatusCode)
	}

	var validateTokenInfo models.TwitchOatValidateTokenhResponse
	err = jsoniter.Unmarshal(readedResp, &validateTokenInfo)
	if err != nil {
		return
	}

	data = &validateTokenInfo

	return
}
