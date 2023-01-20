package notification

import (
	"context"
	"database/sql"
	"strconv"
	"time"
	"twitch_telegram_bot/internal/models"

	"github.com/jmoiron/sqlx"
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

		notifications, err := tn.GetTwitchNotificationsBatch(ctx, lastId)
		if err != nil {
			return errors.Wrap(err, "Sync")
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

				tokens, err := tn.GetTokensByChatID(ctx, notification.ChatId)
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

							err = tn.UpdateChatTokensByState(ctx, tokens.State, newTokens.AccessToken, newTokens.RefreshToken)
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
						for _, streamInfo := range streams.StreamInfo {

							tx, err := tn.db.BeginTxx(ctx, &sql.TxOptions{})
							if err != nil {
								return errors.Wrap(err, "BeginTxx")
							}

							streamIdInt, err := strconv.ParseUint(streamInfo.StreamId, 10, 64)
							if err != nil {
								logrus.Errorf("cannot parse %s to uint64", streamInfo.StreamId)
								tx.Rollback()
								continue
							}

							err = tn.AddTwitchNotificationLog(ctx, tx, streamIdInt, notification.ChatId)
							if err != nil {
								logrus.Errorf("cannot add notification log for %d", streamIdInt)
								tx.Rollback()
								continue
							}

							if currentTime.Before(streamInfo.StartedAt.Add(time.Minute * 10)) {
								err = tn.ThrowNotification(ctx, streamInfo, notification.ChatId)
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

				}

			}

		}

		streams, err := tn.twitchClient.GetActiveStreamInfoByUsers(ctx, users)
		if err != nil {
			return errors.Wrap(err, "GetActiveStreamInfoByUsers")
		}

		if streams != nil {
			for _, notification := range notifications {
				for _, streamInfo := range streams.StreamInfo {
					if notification.TwitchUser == streamInfo.UserId ||
						notification.TwitchUser == streamInfo.UserLogin ||
						notification.TwitchUser == streamInfo.UserName {

						tx, err := tn.db.BeginTxx(ctx, &sql.TxOptions{})
						if err != nil {
							return errors.Wrap(err, "BeginTxx")
						}

						streamIdInt, err := strconv.ParseUint(streamInfo.StreamId, 10, 64)
						if err != nil {
							logrus.Errorf("cannot parse %s to uint64", streamInfo.StreamId)
							tx.Rollback()
							continue
						}

						err = tn.AddTwitchNotificationLog(ctx, tx, streamIdInt, notification.ChatId)
						if err != nil {
							logrus.Errorf("cannot add notification log for %d", streamIdInt)
							tx.Rollback()
							continue
						}

						if currentTime.Before(streamInfo.StartedAt.Add(time.Minute * 10)) {

							err = tn.ThrowNotification(ctx, streamInfo, notification.ChatId)
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
			}

		}

	}
}

type GetTwitchNotificationsResponse struct {
	ID          uint64                        `db:"id"`
	ChatId      uint64                        `db:"chat_id"`
	TwitchUser  string                        `db:"twitch_user"`
	RequestType models.StreamNotificationType `db:"request_type"`
}

const batchSize = 100

func (tn *TwitchNotificationService) GetTwitchNotificationsBatch(ctx context.Context,
	lastId uint64) (notifInfo []GetTwitchNotificationsResponse, err error) {

	query := `
				select 
					tn.id, 
					tn.chat_id, 
					tn.twitch_user,
					tn.request_type 
				from twitch_notifications tn
				where tn.is_active = true 
					and tn.id > $1
				order by tn.id
				limit $2;
			`

	err = tn.db.SelectContext(ctx, &notifInfo, query, lastId, batchSize)
	if err != nil {
		return []GetTwitchNotificationsResponse{}, errors.Wrap(err, "GetTwitchNotificationsBatch selectContext")
	}

	return
}

func (tn *TwitchNotificationService) AddTwitchNotification(ctx context.Context, chatId uint64, user string, notiType models.StreamNotificationType) (err error) {

	query := `
				insert into twitch_notifications (chat_id, twitch_user, request_type) 
					values ($1, $2, $3)
				on conflict (chat_id, twitch_user, request_type) 
					do update
					set (request_type, is_active, updated_at) = ($3, true, now());
	`

	res, err := tn.db.ExecContext(ctx, query, chatId, user, notiType)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return
}

func (tn *TwitchNotificationService) SetInactiveNotificationByUser(ctx context.Context, chatId uint64, user string) (err error) {

	query := `
				update twitch_notifications 
					set (is_active, updated_at) = (false, now())
					where (chat_id, twitch_user, request_type) = ($1, $2, 'by_user');
	`

	res, err := tn.db.ExecContext(ctx, query, chatId, user)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n < 1 {
		return errors.New("notification not found")
	}

	return
}

func (tn *TwitchNotificationService) SetInactiveNotificationFollowed(ctx context.Context, chatId uint64) (err error) {

	query := `
				update twitch_notifications 
					set (is_active, updated_at) = (false, now())
					where (chat_id, request_type) = ($1, 'followed');
	`

	res, err := tn.db.ExecContext(ctx, query, chatId)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n < 1 {
		return errors.New("notification not found")
	}

	return
}

func (tn *TwitchNotificationService) AddTwitchNotificationLog(ctx context.Context, tx *sqlx.Tx, streamId, chatId uint64) (err error) {

	query := `
				insert into twitch_notifications_log (stream_id, chat_id) 
					values ($1, $2)
				on conflict (stream_id, chat_id) do nothing;
	`

	res, err := tx.ExecContext(ctx, query, streamId, chatId)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n < 1 {
		return errors.New("no rows insert")
	}

	return
}

type GetTwitchNotificationLogResponse struct {
	ChatId     uint64 `db:"chat_id"`
	TwitchUser string `db:"twitch_user"`
	StreamId   uint64 `db:"stream_id"`
}

func (tn *TwitchNotificationService) GetTwitchNotificationLogByStreamId(ctx context.Context,
	streamId uint64) (logInfo *GetTwitchNotificationsResponse, err error) {

	query := `
				select 
					tnl.chat_id,  
					tnl.stream_id 
				from twitch_notifications_log tnl 
				where tnl.stream_id = $1;
			`

	err = tn.db.GetContext(ctx, logInfo, query, streamId)
	if err != nil {
		return &GetTwitchNotificationsResponse{}, errors.Wrap(err, "GetTwitchNotificationLogByStreamId getContext")
	}

	return
}

type tokens struct {
	AccessToken  *string `db:"access_token"`
	RefreshToken *string `db:"refresh_token"`
	State        string  `db:"current_state"`
}

func (tn *TwitchNotificationService) GetTokensByChatID(ctx context.Context,
	chatID uint64) (data tokens, err error) {

	query := `
			select 
				tut.access_token, 
				tut.refresh_token, 
				tut.current_state
			from twitch_user_tokens tut 
			where 'user:read:follows' = ANY(tut."scope") 
				and tut.chat_id = $1;
			`

	err = tn.db.GetContext(ctx, &data, query, chatID)

	return
}

func (tn *TwitchNotificationService) UpdateChatTokensByState(ctx context.Context, state, accessToken, refreshToken string) (err error) {

	query := `
		update twitch_user_tokens 
			set (access_token, refresh_token, updated_at) = ($1, $2, now())
		where current_state = $3;
	`
	_, err = tn.db.ExecContext(ctx, query, accessToken, refreshToken, state)
	if err != nil {
		return err
	}

	return
}
