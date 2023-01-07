package twitch_user_authorization

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

const (
	twitchIDSchemeHost string = "https://id.twitch.tv"
	tokenInvalid       string = "token invalid"
)

type CheckUserTokensByChatResp struct {
	AccessToken string
	UserID      string
	Link        string
}

func (tuas *TwitchUserAuthorizationService) CheckUserTokensByChat(ctx context.Context, chatID uint64) (CheckUserTokensByChatResp, error) {

	data, err := tuas.GetTokensByChat(ctx, chatID)
	if err != nil {
		return CheckUserTokensByChatResp{}, err
	}
	if data == nil {

		st := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprint(chatID))))

		err := tuas.AddChatInfo(ctx, chatID, st)
		if err != nil {
			return CheckUserTokensByChatResp{}, nil
		}

		link := tuas.TwitchCreateOAuth2Link(ctx, st)

		return CheckUserTokensByChatResp{
			Link: link,
		}, nil
	}

	if data.AccessToken == nil {
		return CheckUserTokensByChatResp{
			Link: tuas.TwitchCreateOAuth2Link(ctx, data.CurrentState),
		}, nil
	}

	validData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, *data.AccessToken)

	if err != nil {
		if err.Error() == tokenInvalid {

			newTokens, err := tuas.twitchOauthClient.TwitchGetUserTokenRefresh(ctx, *data.RefreshToken)
			if err != nil {
				return CheckUserTokensByChatResp{}, errors.Wrap(err, "TwitchGetUserTokenRefresh")
			}

			err = tuas.UpdateChatTokens(ctx, chatID, newTokens.AccessToken, newTokens.RefreshToken)
			if err != nil {
				return CheckUserTokensByChatResp{}, errors.Wrap(err, "UpdateChatTokens")
			}

			return CheckUserTokensByChatResp{
				AccessToken: newTokens.AccessToken,
				UserID:      validData.UserId,
			}, nil

		}
		return CheckUserTokensByChatResp{}, errors.Wrap(err, "TwitchOAuthGetToken")
	}

	return CheckUserTokensByChatResp{
		AccessToken: *data.AccessToken,
		UserID:      validData.UserId,
	}, nil
}

type GetTokensByChatResp struct {
	AccessToken  *string
	RefreshToken *string
	CurrentState string
}

func (tuas *TwitchUserAuthorizationService) GetTokensByChat(ctx context.Context, chatID uint64) (data *GetTokensByChatResp, err error) {

	query := `
		select 
			access_token, 
			refresh_token, 
			current_state 
		from twitch_user_tokens tut 
		where chat_id = $1;
	`

	err = tuas.db.GetContext(ctx, &data, query)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return
}

func (tuas *TwitchUserAuthorizationService) AddChatInfo(ctx context.Context, chatID uint64, state string) (err error) {

	query := `
		insert into twitch_user_tokens (chat_id, current_state) 
			values ($1, $2)
		on conflict do nothing;
	`
	_, err = tuas.db.ExecContext(ctx, query, chatID, state)

	return
}

func (tuas *TwitchUserAuthorizationService) UpdateChatTokens(ctx context.Context, chatID uint64, accessToken string, refreshToken string) (err error) {

	query := `
		update twitch_user_tokens 
			set (access_token, refresh_token) = ($1, $2)
		where chat_id = $3;
	`
	_, err = tuas.db.ExecContext(ctx, query, accessToken, refreshToken, chatID)
	if err != nil {
		return err
	}

	return
}

func (tuas *TwitchUserAuthorizationService) TwitchCreateOAuth2Link(ctx context.Context, state string) string {

	url := fmt.Sprintf("%s/oauth2/authorize?client_id=%s&response_type=token&redirect_uri=http://localhost:3000&scope=user:read:follows&state=%s",
		twitchIDSchemeHost, os.Getenv("TWITCH_CLIENT_ID"), state)

	return url
}
