package twitch_client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"
	"twitch_telegram_bot/internal/models"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var digitCheck = regexp.MustCompile(`^[0-9]+$`) // check if have only digits

// TODO: запрос не возвращает инфу по несуществующему пользователю,
// сделать пост проверку на точное соответсвие айдишников на входе и выходе
func (twc *TwitchClient) GetUserInfo(ctx context.Context, token string, ids []string) (data *models.GetUserInfoResponse, err error) {

	client := http.Client{
		Timeout: time.Second * 5,
	}

	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

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

		return nil, errors.Errorf("get twitch users failed with status code: %d", resp.StatusCode)
	}

	readedResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var usersInfo models.GetUserInfoResponse
	err = jsoniter.Unmarshal(readedResp, &usersInfo)
	if err != nil {
		return
	}

	data = &usersInfo

	return
}
