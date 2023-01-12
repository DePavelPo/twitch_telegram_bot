package twitch_user_authorization

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	twitchIDSchemeHost  string = "https://id.twitch.tv"
	tokenInvalid        string = "token invalid"
	refreshTokenInvalid string = "Invalid refresh token"
)

type CheckUserTokensByChatResp struct {
	AccessToken string
	UserID      string
	Link        string
}

func (tuas *TwitchUserAuthorizationService) CheckUserTokensByChat(ctx context.Context, chatID uint64) (CheckUserTokensByChatResp, error) {

	data, err := tuas.GetTokensByChat(ctx, chatID)
	if err != nil {

		if err == sql.ErrNoRows {

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

		return CheckUserTokensByChatResp{}, errors.Wrap(err, "GetTokensByChat")
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

				if err.Error() == refreshTokenInvalid {

					return CheckUserTokensByChatResp{
						Link: tuas.TwitchCreateOAuth2Link(ctx, data.CurrentState),
					}, nil

				}

				return CheckUserTokensByChatResp{}, errors.Wrap(err, "TwitchGetUserTokenRefresh")
			}

			newValidData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, newTokens.AccessToken)
			if err != nil {
				return CheckUserTokensByChatResp{}, errors.Wrap(err, "TwitchOAuthValidateToken")
			}

			err = tuas.UpdateChatTokens(ctx, chatID, newTokens.AccessToken, newTokens.RefreshToken)
			if err != nil {
				return CheckUserTokensByChatResp{}, errors.Wrap(err, "UpdateChatTokens")
			}

			return CheckUserTokensByChatResp{
				AccessToken: newTokens.AccessToken,
				UserID:      newValidData.UserId,
			}, nil

		}
		return CheckUserTokensByChatResp{}, errors.Wrap(err, "TwitchOAuthValidateToken")
	}

	return CheckUserTokensByChatResp{
		AccessToken: *data.AccessToken,
		UserID:      validData.UserId,
	}, nil
}

type GetTokensByChatResp struct {
	AccessToken  *string `db:"access_token"`
	RefreshToken *string `db:"refresh_token"`
	CurrentState string  `db:"current_state"`
}

func (tuas *TwitchUserAuthorizationService) GetTokensByChat(ctx context.Context, chatID uint64) (data GetTokensByChatResp, err error) {

	query := `
		select 
			access_token, 
			refresh_token, 
			current_state 
		from twitch_user_tokens tut 
		where chat_id = $1;
	`

	err = tuas.db.GetContext(ctx, &data, query, chatID)

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

	url := fmt.Sprintf(
		"%s/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=http://localhost:3000&scope=user:read:follows&state=%s",
		twitchIDSchemeHost,
		os.Getenv("TWITCH_CLIENT_ID"),
		state,
	)

	return url
}

func (tuas *TwitchUserAuthorizationService) CheckUserTokensByState(ctx context.Context, code, state string) error {

	data, err := tuas.GetTokensByState(ctx, state)

	if err != nil {

		if err == sql.ErrNoRows {
			logrus.Error("GetTokensByState error: state not found")
		}

		return errors.Wrap(err, "GetTokensByState")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		return err
	}

	if data.AccessToken == nil {

		tokens, err := tuas.twitchOauthClient.TwitchGetUserToken(ctx, code)
		if err != nil {
			return errors.Wrap(err, "TwitchGetUserToken")
		}

		err = tuas.UpdateChatTokensByState(ctx, state, tokens.AccessToken, tokens.RefreshToken)
		if err != nil {
			return errors.Wrap(err, "UpdateChatTokensByState")
		}

		msg := tgbotapi.NewMessage(int64(data.ChatID), "")

		resp := "Sorry, the functional is not available now"
		msg.Text = resp

		_, err = bot.Send(msg)
		if err != nil {
			logrus.Infof("telegram send message error: %v", err)
		}

		return nil

	}

	_, err = tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, *data.AccessToken)
	if err != nil {
		if err.Error() == tokenInvalid {

			newTokens, err := tuas.twitchOauthClient.TwitchGetUserTokenRefresh(ctx, *data.RefreshToken)
			if err != nil {

				if err.Error() == refreshTokenInvalid {

					tokens, err := tuas.twitchOauthClient.TwitchGetUserToken(ctx, code)
					if err != nil {
						return errors.Wrap(err, "TwitchGetUserToken")
					}

					err = tuas.UpdateChatTokensByState(ctx, state, tokens.AccessToken, tokens.RefreshToken)
					if err != nil {
						return errors.Wrap(err, "UpdateChatTokensByState")
					}

					msg := tgbotapi.NewMessage(int64(data.ChatID), "")

					resp := "Sorry, the functional is not available now"
					msg.Text = resp

					_, err = bot.Send(msg)
					if err != nil {
						logrus.Infof("telegram send message error: %v", err)
					}

					return nil

				}

				return errors.Wrap(err, "TwitchGetUserTokenRefresh")
			}

			_, err = tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, newTokens.AccessToken)
			if err != nil {
				return errors.Wrap(err, "TwitchOAuthValidateToken")
			}

			err = tuas.UpdateChatTokensByState(ctx, state, newTokens.AccessToken, newTokens.RefreshToken)
			if err != nil {
				return errors.Wrap(err, "UpdateChatTokensByState")
			}

			msg := tgbotapi.NewMessage(int64(data.ChatID), "")

			resp := "Sorry, the functional is not available now"
			msg.Text = resp

			_, err = bot.Send(msg)
			if err != nil {
				logrus.Infof("telegram send message error: %v", err)
			}

			return nil

		}
		return errors.Wrap(err, "TwitchOAuthValidateToken")
	}

	msg := tgbotapi.NewMessage(int64(data.ChatID), "")

	resp := "Sorry, the functional is not available now"
	msg.Text = resp

	_, err = bot.Send(msg)
	if err != nil {
		logrus.Infof("telegram send message error: %v", err)
	}

	return nil
}

type GetTokensByStateResp struct {
	AccessToken  *string `db:"access_token"`
	RefreshToken *string `db:"refresh_token"`
	ChatID       uint64  `db:"chat_id"`
}

func (tuas *TwitchUserAuthorizationService) GetTokensByState(ctx context.Context, state string) (data GetTokensByStateResp, err error) {

	query := `
		select 
			chat_id,
			access_token, 
			refresh_token 
		from twitch_user_tokens tut 
		where current_state = $1;
	`

	err = tuas.db.GetContext(ctx, &data, query, state)

	return
}

func (tuas *TwitchUserAuthorizationService) UpdateChatTokensByState(ctx context.Context, state, accessToken, refreshToken string) (err error) {

	query := `
		update twitch_user_tokens 
			set (access_token, refresh_token) = ($1, $2)
		where current_state = $3;
	`
	_, err = tuas.db.ExecContext(ctx, query, accessToken, refreshToken, state)
	if err != nil {
		return err
	}

	return
}
