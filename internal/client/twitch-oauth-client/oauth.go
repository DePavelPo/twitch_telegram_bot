package twitch_oath_client

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

const twitchIDSchemeHost string = "https://id.twitch.tv"

type TwitchOauthClient struct {
	protocol     string
	redirectAddr string
}

func NewTwitchOauthClient(
	protocol, redirectAddr string,
) *TwitchOauthClient {
	return &TwitchOauthClient{
		protocol:     protocol,
		redirectAddr: redirectAddr,
	}
}

func (twc *TwitchOauthClient) TwitchOAuthGetToken(ctx context.Context) (data *models.TwitchOautGetTokenResponse, err error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("POST", twitchIDSchemeHost+"/oauth2/token", nil)
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

	var tokenInfo models.TwitchOautGetTokenResponse
	err = jsoniter.Unmarshal(readedResp, &tokenInfo)
	if err != nil {
		return
	}

	data = &tokenInfo

	return
}

func (twc *TwitchOauthClient) TwitchOAuthValidateToken(ctx context.Context, token string) (data *models.TwitchOautValidateTokenResponse, err error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", twitchIDSchemeHost+"/oauth2/validate", nil)
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

		return nil, errors.Errorf("validate token failed with status code: %d", resp.StatusCode)
	}

	var validateTokenInfo models.TwitchOautValidateTokenResponse
	err = jsoniter.Unmarshal(readedResp, &validateTokenInfo)
	if err != nil {
		return
	}

	data = &validateTokenInfo

	return
}
