package twitch_oath_client

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

func (twc *TwitchOauthClient) TwitchOAuth2(ctx context.Context) (err error) {

	// client := http.Client{
	// 	Timeout: time.Second * 5,
	// }

	// req, err := http.NewRequest("GET", twitchIdSchemeHost+"/oauth2/authorize", nil)
	// if err != nil {
	// 	return err
	// }

	// query := req.URL.Query()
	// query.Add("client_id", os.Getenv("TWITCH_CLIENT_ID"))
	// query.Add("response_type", "token")
	// query.Add("redirect_uri", "http://localhost:3000")
	// query.Add("scope", "user:read:follows")
	// query.Add("state", "c3ab8aa609ea11e793be92451f192671")
	// req.URL.RawQuery = query.Encode()
	url := "https://id.twitch.tv/oauth2/authorize?client_id=9ktj5w1ir11yf4s5t7guo1eku6i6vx&response_type=token&redirect_uri=http://localhost:3000&scope=user:read:follows&state=c3ab8aa609ea11e793ae92361f192671"
	url = strings.ReplaceAll(url, "&", "^&")
	logrus.Info(url)
	exec.Command("cmd", "/c", "start", url).Run()

	// resp, err := client.Do(req)
	// if err != nil {
	// 	return
	// }

	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return errors.Errorf("error with status code %d", resp.StatusCode)
	// }

	return nil
}

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
	query.Add("redirect_uri", "http://localhost:3000")
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

	var tokenInfo models.TwitchOautGetUserTokenResponse
	err = jsoniter.Unmarshal(readedResp, &tokenInfo)
	if err != nil {
		return
	}

	data = &tokenInfo

	return
}
