package twitch_oath_client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

func (twc *TwitchOauthClient) TwitchOAuthGetToken(
	ctx context.Context,
	twitchCode string,
) (*models.TwitchOautGetTokenResponse, error) {

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

	// client token flow as a default
	grantType := "client_credentials"

	// if there is twitch code - it's user token flow
	if twitchCode != "" {
		grantType = "authorization_code"

		query.Add("code", twitchCode)
		query.Add("redirect_uri", fmt.Sprintf("%s://%s", twc.protocol, twc.redirectAddr))
	}
	query.Add("grant_type", grantType)

	req.URL.RawQuery = query.Encode()

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if expectableErrorCode[resp.StatusCode] {

			readedResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var errorResp models.ValidateTokenInvalid
			err = jsoniter.Unmarshal(readedResp, &errorResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New(errorResp.Message)
		}

		return nil, errors.Errorf("[%s] get token failed with status code: %d", grantType, resp.StatusCode)
	}

	readedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.TwitchOautGetTokenResponse
	err = jsoniter.Unmarshal(readedResp, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (twc *TwitchOauthClient) TwitchGetUserTokenRefresh(
	ctx context.Context,
	token string,
) (*models.TwitchOautGetTokenResponse, error) {

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
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if expectableErrorCode[resp.StatusCode] {

			readedResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var errorResp models.ValidateTokenInvalid
			err = jsoniter.Unmarshal(readedResp, &errorResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New(errorResp.Message)
		}

		return nil, errors.Errorf("refresh token failed with status code: %d", resp.StatusCode)
	}

	readedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.TwitchOautGetTokenResponse
	err = jsoniter.Unmarshal(readedResp, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (twc *TwitchOauthClient) TwitchOAuthValidateToken(
	ctx context.Context,
	token string,
) (*models.TwitchOautValidateTokenResponse, error) {

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
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if expectableErrorCode[resp.StatusCode] {

			readedResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var errorResp models.ValidateTokenInvalid
			err = jsoniter.Unmarshal(readedResp, &errorResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New(errorResp.Message)
		}

		return nil, errors.Errorf("validate token failed with status code: %d", resp.StatusCode)
	}

	readedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.TwitchOautValidateTokenResponse
	err = jsoniter.Unmarshal(readedResp, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
