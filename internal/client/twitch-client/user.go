package twitch_client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var digitCheck = regexp.MustCompile(`^[0-9]+$`) // check if have only digits

func (twc *TwitchClient) GetUserInfo(ctx context.Context, ids []string) (*models.GetUserInfoResponse, error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", models.TwitchApiSchemeHost+"/helix/users", nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for _, id := range ids {
		if digitCheck.MatchString(id) {
			query.Add("id", id)
			continue
		}
		query.Add("login", id)
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

		return nil, errors.Errorf("get twitch users failed with status code: %d", resp.StatusCode)
	}

	readedResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response models.GetUserInfoResponse
	err = jsoniter.Unmarshal(readedResp, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
