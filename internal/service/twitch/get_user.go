package twitch_service

import (
	"context"
	"twitch_telegram_bot/internal/models"

	"github.com/pkg/errors"
)

func (tws *TwitchService) GetUser(ctx context.Context, token, id string) (*models.GetUserInfoResponse, error) {

	usersStruct := []string{id}

	userInfo, err := tws.twitchClient.GetUserInfo(ctx, token, usersStruct)
	if err != nil {
		return nil, err
	}

	if userInfo == nil {
		return userInfo, errors.New("empty response stuct")
	}

	if userInfo.Data[0].UserID != id && userInfo.Data[0].Login != id && userInfo.Data[0].DisplayName != id {
		return nil, errors.Errorf("invalid reponse data, give %s, got id %s, login %s, name %s",
			id, userInfo.Data[0].UserID, userInfo.Data[0].Login, userInfo.Data[0].DisplayName)
	}

	return userInfo, nil

}
