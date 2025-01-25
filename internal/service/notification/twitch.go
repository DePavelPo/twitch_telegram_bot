package notification

import (
	"context"
	"slices"
	"strconv"
	"time"

	"twitch_telegram_bot/internal/models"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	twitchNotificationBGSync = "twitchNotification_BGSync"
)

func (tn *TwitchNotificationService) Sync(ctx context.Context) error {

	var (
		lastId      uint64    = 0
		currentTime time.Time = time.Now()
	)

	for {

		notifications, err := tn.dbRepo.GetTwitchNotificationsBatch(ctx, lastId)
		if err != nil {
			return errors.Wrap(err, "GetTwitchNotificationsBatch")
		}

		if len(notifications) == 0 {
			return nil
		}

		lastId = notifications[len(notifications)-1].ID

		users := []string{}

		for _, notification := range notifications {

			switch notification.RequestType {
			case models.NotificationByUser:
				users = append(users, notification.TwitchUser)
			case models.NotificationFollowed:

				tokens, err := tn.dbRepo.GetTokensByChatID(ctx, notification.ChatId, models.UserReadFollows)
				if err != nil {
					return errors.Wrap(err, "GetTokensByChatID")
				}

				if tokens.AccessToken != nil {

					var streams *models.Streams

					streams, err = tn.twitchClient.GetActiveFollowedStreams(ctx, notification.TwitchUser, *tokens.AccessToken)
					if err != nil {

						if err.Error() == models.InvalidOathToken {

							newTokens, err := tn.twitchOauthClient.TwitchGetUserTokenRefresh(ctx, *tokens.RefreshToken)
							if err != nil {

								if err.Error() == models.RefreshTokenInvalid {

									logrus.Errorf("Sync notification error: %s", models.RefreshTokenInvalid)
									break

								}
							}

							err = tn.dbRepo.UpdateChatTokensByState(ctx, tokens.State, newTokens.AccessToken, newTokens.RefreshToken)
							if err != nil {

								logrus.Errorf("Sync notification UpdateChatTokensByState error: %v", err)
								break

							}

							streams, err = tn.twitchClient.GetActiveFollowedStreams(ctx, notification.TwitchUser, newTokens.AccessToken)
							if err != nil {
								logrus.Errorf("Sync notification GetActiveFollowedStreams error: %v", err)
								break
							}

						} else {

							return errors.Wrap(err, "GetActiveFollowedStreams")

						}

					}

					if streams != nil {

						err = tn.LogAndThrowNotifications(ctx, streams.StreamInfo, notification.TwitchUser, notification.ChatId, currentTime)
						if err != nil {
							return errors.Wrap(err, "LogAndThrowNotifications")
						}

					}

				}

			}

		}

		streams, err := tn.twitchClient.GetActiveStreamInfoByUsers(ctx, users)
		if err != nil {
			return errors.Wrap(err, "GetActiveStreamInfoByUsers")
		}

		if streams != nil {
			for _, notification := range notifications {
				err = tn.LogAndThrowNotifications(ctx, streams.StreamInfo, notification.TwitchUser, notification.ChatId, currentTime)
				if err != nil {
					return errors.Wrap(err, "LogAndThrowNotifications")
				}
			}

		}

	}
}

// log info about the notification to db and throw the notification to the bot
func (tn *TwitchNotificationService) LogAndThrowNotifications(
	ctx context.Context,
	streamsInfo []models.Stream,
	twitchUser string,
	chatID uint64,
	currentTime time.Time,
) error {

	for _, streamInfo := range streamsInfo {
		if slices.Contains([]string{streamInfo.UserId, streamInfo.UserLogin, streamInfo.UserName}, twitchUser) {

			tx, err := tn.dbRepo.BeginTransaction(ctx)
			if err != nil {
				return errors.Wrap(err, "BeginTransaction")
			}

			streamIdInt, err := strconv.ParseUint(streamInfo.StreamId, 10, 64)
			if err != nil {
				logrus.Errorf("cannot parse %s to uint64", streamInfo.StreamId)
				tx.Rollback()
				continue
			}

			err = tn.dbRepo.AddTwitchNotificationLog(ctx, tx, streamIdInt, chatID)
			if err != nil {
				logrus.Errorf("cannot add notification log for %d", streamIdInt)
				tx.Rollback()
				continue
			}

			// throw notifications for streams
			// that started less then 10 minutes ago
			if currentTime.Before(streamInfo.StartedAt.Add(time.Minute * 10)) {

				err = tn.ThrowNotification(ctx, streamInfo, chatID)
				if err != nil {
					tx.Rollback()
					return errors.Wrap(err, "ThrowNotification")
				}

			}

			if err = tx.Commit(); err != nil {
				return errors.Wrap(err, "Commit")
			}

		}

	}

	return nil
}
