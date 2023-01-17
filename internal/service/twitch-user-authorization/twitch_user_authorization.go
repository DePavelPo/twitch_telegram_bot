package twitch_user_authorization

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"twitch_telegram_bot/internal/models"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	twitchIDSchemeHost string = "https://id.twitch.tv"
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
		if err.Error() == models.TokenInvalid {

			newTokens, err := tuas.twitchOauthClient.TwitchGetUserTokenRefresh(ctx, *data.RefreshToken)
			if err != nil {

				if err.Error() == models.RefreshTokenInvalid {

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
		"%s/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=http://localhost:3000&scope=user:read:follows+channel:read:subscriptions&state=%s",
		twitchIDSchemeHost,
		os.Getenv("TWITCH_CLIENT_ID"),
		state,
	)

	return url
}

func (tuas *TwitchUserAuthorizationService) CheckUserTokensByState(ctx context.Context, code, state string, scope []string) error {

	data, err := tuas.UpdateScopeAndGetTokens(ctx, scope, state)

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

		err = tuas.AddTwitchNotification(ctx, data.ChatID, "", models.NotificationFollowed)
		if err != nil {
			return errors.Wrap(err, "AddTwitchNotification")
		}

		msg := tgbotapi.NewMessage(int64(data.ChatID), "")

		resp := "Request successfully accepted! This channel will now receive stream notifications from channels that you following"
		msg.Text = resp

		_, err = bot.Send(msg)
		if err != nil {
			logrus.Infof("telegram send message error: %v", err)
		}

		return nil

	}

	_, err = tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, *data.AccessToken)
	if err != nil {
		if err.Error() == models.TokenInvalid {

			newTokens, err := tuas.twitchOauthClient.TwitchGetUserTokenRefresh(ctx, *data.RefreshToken)
			if err != nil {

				if err.Error() == models.RefreshTokenInvalid {

					tokens, err := tuas.twitchOauthClient.TwitchGetUserToken(ctx, code)
					if err != nil {
						return errors.Wrap(err, "TwitchGetUserToken")
					}

					err = tuas.UpdateChatTokensByState(ctx, state, tokens.AccessToken, tokens.RefreshToken)
					if err != nil {
						return errors.Wrap(err, "UpdateChatTokensByState")
					}

					err = tuas.AddTwitchNotification(ctx, data.ChatID, "", models.NotificationFollowed)
					if err != nil {
						return errors.Wrap(err, "AddTwitchNotification")
					}

					msg := tgbotapi.NewMessage(int64(data.ChatID), "")

					resp := "Request successfully accepted! This channel will now receive stream notifications from channels that you following"
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

			err = tuas.AddTwitchNotification(ctx, data.ChatID, "", models.NotificationFollowed)
			if err != nil {
				return errors.Wrap(err, "AddTwitchNotification")
			}

			msg := tgbotapi.NewMessage(int64(data.ChatID), "")

			resp := "Request successfully accepted! This channel will now receive stream notifications from channels that you following"
			msg.Text = resp

			_, err = bot.Send(msg)
			if err != nil {
				logrus.Infof("telegram send message error: %v", err)
			}

			return nil

		}
		return errors.Wrap(err, "TwitchOAuthValidateToken")
	}

	err = tuas.AddTwitchNotification(ctx, data.ChatID, "", models.NotificationFollowed)
	if err != nil {
		return errors.Wrap(err, "AddTwitchNotification")
	}

	msg := tgbotapi.NewMessage(int64(data.ChatID), "")

	resp := "Request successfully accepted! This channel will now receive stream notifications from channels that you following"
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

func (tuas *TwitchUserAuthorizationService) UpdateScopeAndGetTokens(ctx context.Context, scope []string, state string) (data GetTokensByStateResp, err error) {

	query := `
		update twitch_user_tokens
			set scope = $1
			where current_state = $2
			returning 
				chat_id,
				access_token, 
				refresh_token;
	`

	err = tuas.db.GetContext(ctx, &data, query, pq.StringArray(scope), state)

	return
}

func (tuas *TwitchUserAuthorizationService) UpdateChatTokensByState(ctx context.Context, state, accessToken, refreshToken string) (err error) {

	query := `
		update twitch_user_tokens 
			set (access_token, refresh_token, updated_at) = ($1, $2, now())
		where current_state = $3;
	`
	_, err = tuas.db.ExecContext(ctx, query, accessToken, refreshToken, state)
	if err != nil {
		return err
	}

	return
}

func (tuas *TwitchUserAuthorizationService) AddTwitchNotification(ctx context.Context, chatId uint64, user string, notiType models.StreamNotificationType) (err error) {

	query := `
				insert into twitch_notifications (chat_id, twitch_user, request_type) 
					values ($1, $2, $3)
				on conflict (chat_id, twitch_user, request_type) 
					do update
					set (request_type, is_active) = ($3, true);
	`

	res, err := tuas.db.ExecContext(ctx, query, chatId, user, notiType)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return
}