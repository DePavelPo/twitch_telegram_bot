package twitch_service

import (
	"context"
	"twitch_telegram_bot/internal/models"

	"github.com/pkg/errors"
)

func (tws *TwitchService) GetActiveStreamInfoByUser(ctx context.Context, id string) (*models.Streams, error) {
	usersStruct := []string{id}

	streamInfo, err := tws.twitchClient.GetActiveStreamInfoByUsers(ctx, usersStruct)
	if err != nil {
		return nil, err
	}

	if streamInfo == nil {
		return streamInfo, errors.New("empty response struct")
	}

	if len(streamInfo.StreamInfo) < 1 {
		return streamInfo, errors.New("stream not found")
	}

	if streamInfo.StreamInfo[0].UserId != id && streamInfo.StreamInfo[0].UserLogin != id && streamInfo.StreamInfo[0].UserName != id {
		return nil, errors.Errorf("invalid response data, give %s, got id %s, login %s, name %s",
			id, streamInfo.StreamInfo[0].UserId, streamInfo.StreamInfo[0].UserLogin, streamInfo.StreamInfo[0].UserName)
	}

	return streamInfo, nil
}
