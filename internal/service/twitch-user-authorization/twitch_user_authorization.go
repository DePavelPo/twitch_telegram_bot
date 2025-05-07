package twitch_user_authorization

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"twitch_telegram_bot/internal/models"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CheckUserTokensByChatResp struct {
	AccessToken string
	UserID      string
	Link        string
}

func (tuas *TwitchUserAuthorizationService) CheckUserTokensByChat(ctx context.Context, chatID uint64) (CheckUserTokensByChatResp, error) {
	data, err := tuas.dbRepo.GetTokensByChat(ctx, chatID)
	if err != nil {

		if err == sql.ErrNoRows {

			st := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprint(chatID))))

			err := tuas.dbRepo.AddChatInfo(ctx, chatID, st)
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
			Link: tuas.TwitchCreateOAuth2Link(ctx, data.State),
		}, nil
	}

	validData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, *data.AccessToken)

	if err != nil {
		if err.Error() == models.TokenInvalid {

			newTokens, err := tuas.twitchOauthClient.TwitchGetUserTokenRefresh(ctx, *data.RefreshToken)
			if err != nil {

				if err.Error() == models.RefreshTokenInvalid {

					return CheckUserTokensByChatResp{
						Link: tuas.TwitchCreateOAuth2Link(ctx, data.State),
					}, nil

				}

				return CheckUserTokensByChatResp{}, errors.Wrap(err, "TwitchGetUserTokenRefresh")
			}

			newValidData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, newTokens.AccessToken)
			if err != nil {
				return CheckUserTokensByChatResp{}, errors.Wrap(err, "TwitchOAuthValidateToken")
			}

			err = tuas.dbRepo.UpdateChatTokens(ctx, chatID, newTokens.AccessToken, newTokens.RefreshToken)
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

func (tuas *TwitchUserAuthorizationService) TwitchCreateOAuth2Link(ctx context.Context, state string) string {

	url := fmt.Sprintf(
		"%s/oauth2/authorize?client_id=%s&response_type=code&redirect_uri=%s://%s&scope=user:read:follows+channel:read:subscriptions&state=%s",
		models.TwitchIDSchemeHost,
		os.Getenv("TWITCH_CLIENT_ID"),
		tuas.protocol,
		tuas.redirectAddr,
		state,
	)

	return url
}

func (tuas *TwitchUserAuthorizationService) CheckUserTokensByState(ctx context.Context, code, state string, scope []string) error {
	data, err := tuas.dbRepo.UpdateScopeAndGetTokens(ctx, scope, state)

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
		tokens, err := tuas.twitchOauthClient.TwitchOAuthGetToken(ctx, code)
		if err != nil {
			return errors.Wrap(err, "get twitch user token")
		}

		err = tuas.dbRepo.UpdateChatTokensByState(ctx, state, tokens.AccessToken, tokens.RefreshToken)
		if err != nil {
			return errors.Wrap(err, "UpdateChatTokensByState")
		}

		validData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, tokens.AccessToken)
		if err != nil {
			return errors.Wrap(err, "TwitchOAuthValidateToken")
		}

		err = tuas.dbRepo.AddTwitchNotification(ctx, data.ChatID, validData.UserId, models.NotificationFollowed)
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

	validData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, *data.AccessToken)
	if err != nil {
		if err.Error() == models.TokenInvalid {

			newTokens, err := tuas.twitchOauthClient.TwitchGetUserTokenRefresh(ctx, *data.RefreshToken)
			if err != nil {
				if err.Error() == models.RefreshTokenInvalid {

					tokens, err := tuas.twitchOauthClient.TwitchOAuthGetToken(ctx, code)
					if err != nil {
						return errors.Wrap(err, "get twitch user token")
					}

					validData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, tokens.AccessToken)
					if err != nil {
						return errors.Wrap(err, "TwitchOAuthValidateToken")
					}

					err = tuas.dbRepo.UpdateChatTokensByState(ctx, state, tokens.AccessToken, tokens.RefreshToken)
					if err != nil {
						return errors.Wrap(err, "UpdateChatTokensByState")
					}

					err = tuas.dbRepo.AddTwitchNotification(ctx, data.ChatID, validData.UserId, models.NotificationFollowed)
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

			validData, err := tuas.twitchOauthClient.TwitchOAuthValidateToken(ctx, newTokens.AccessToken)
			if err != nil {
				return errors.Wrap(err, "TwitchOAuthValidateToken")
			}

			err = tuas.dbRepo.UpdateChatTokensByState(ctx, state, newTokens.AccessToken, newTokens.RefreshToken)
			if err != nil {
				return errors.Wrap(err, "UpdateChatTokensByState")
			}

			err = tuas.dbRepo.AddTwitchNotification(ctx, data.ChatID, validData.UserId, models.NotificationFollowed)
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

	err = tuas.dbRepo.AddTwitchNotification(ctx, data.ChatID, validData.UserId, models.NotificationFollowed)
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
