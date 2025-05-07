package twitch_client

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

	twitch_token_service "twitch_telegram_bot/internal/service/twitch_token"
)

type TwitchClient struct {
	twitchTokenService *twitch_token_service.TwitchTokenService
}

func NewTwitchClient(twitchTokenService *twitch_token_service.TwitchTokenService) *TwitchClient {
	return &TwitchClient{
		twitchTokenService: twitchTokenService,
	}
}

func (twc *TwitchClient) GetActiveStreamInfoByUsers(ctx context.Context, ids []string) (*models.Streams, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", models.TwitchApiSchemeHost+"/helix/streams", nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
			readedResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var errorResp models.TwitchError
			err = jsoniter.Unmarshal(readedResp, &errorResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New(errorResp.Message)
		}

		return nil, errors.Errorf("get twitch streams failed with status code: %d", resp.StatusCode)
	}

	readedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var streamsInfo models.Streams
	err = jsoniter.Unmarshal(readedResp, &streamsInfo)
	if err != nil {
		return nil, err
	}

	return &streamsInfo, nil
}

func (twc *TwitchClient) GetActiveFollowedStreams(ctx context.Context, userID, token string) (*models.Streams, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", models.TwitchApiSchemeHost+"/helix/streams/followed", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("user_id", userID)
	req.URL.RawQuery = query.Encode()

	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			readedResp, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var unauthorizedResp models.TwitchError
			err = jsoniter.Unmarshal(readedResp, &unauthorizedResp)
			if err != nil {
				return nil, err
			}

			return nil, errors.New(unauthorizedResp.Message)
		}

		return nil, errors.Errorf("get twitch streams failed with status code: %d", resp.StatusCode)
	}

	readedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var streamsInfo models.Streams
	err = jsoniter.Unmarshal(readedResp, &streamsInfo)
	if err != nil {
		return nil, err
	}

	return &streamsInfo, nil
}
