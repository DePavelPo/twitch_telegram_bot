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

	twitch_token_service "twitch_telegram_bot/internal/service/twitch_token"
)

const twitchApiSchemeHost string = "https://api.twitch.tv"

type TwitchClient struct {
	twitchTokenService *twitch_token_service.TwitchTokenService
}

func NewTwitchClient(twitchTokenService *twitch_token_service.TwitchTokenService) *TwitchClient {
	return &TwitchClient{
		twitchTokenService: twitchTokenService,
	}
}

func (twc *TwitchClient) GetActiveStreamInfoByUsers(ctx context.Context, ids []string) (data *models.Streams, err error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", twitchApiSchemeHost+"/helix/streams", nil)
	if err != nil {
		return
	}

	query := req.URL.Query()
	for _, id := range ids {
		if digitCheck.MatchString(id) {
			query.Add("user_id", id)
			continue
		}
		query.Add("user_login", id)
	}
	req.URL.RawQuery = query.Encode()

	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twc.twitchTokenService.GetCurrentToken(ctx)))

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			readedResp, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var unauthorizedResp models.GetUserUnauthorized
			err = jsoniter.Unmarshal(readedResp, &unauthorizedResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New(unauthorizedResp.Message)
		}

		return nil, errors.Errorf("get twitch streams failed with status code: %d", resp.StatusCode)
	}

	readedResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var streamsInfo models.Streams
	err = jsoniter.Unmarshal(readedResp, &streamsInfo)
	if err != nil {
		return
	}

	data = &streamsInfo

	return
}
