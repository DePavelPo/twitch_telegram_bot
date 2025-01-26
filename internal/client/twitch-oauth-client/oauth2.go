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

func (twc *TwitchOauthClient) TwitchGetUserToken(ctx context.Context, token string) (data *models.TwitchOautGetUserTokenResponse, err error) {

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
	query.Add("grant_type", "authorization_code")
	query.Add("code", token)
	query.Add("redirect_uri", fmt.Sprintf("%s://%s", twc.protocol, twc.redirectAddr))
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

	var tokenInfo models.TwitchOautGetUserTokenResponse
	err = jsoniter.Unmarshal(readedResp, &tokenInfo)
	if err != nil {
		return
	}

	data = &tokenInfo

	return
}

func (twc *TwitchOauthClient) TwitchGetUserTokenRefresh(ctx context.Context, token string) (data *models.TwitchOautGetUserTokenResponse, err error) {

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
	query.Add("grant_type", "refresh_token")
	query.Add("refresh_token", token)
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

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {

			var invalidTokenResp models.ValidateTokenInvalid
			err = jsoniter.Unmarshal(readedResp, &invalidTokenResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New("Invalid refresh token")
		}

		return nil, errors.Errorf("refresh token failed with status code: %d", resp.StatusCode)
	}

	var tokenInfo models.TwitchOautGetUserTokenResponse
	err = jsoniter.Unmarshal(readedResp, &tokenInfo)
	if err != nil {
		return
	}

	data = &tokenInfo

	return
}
