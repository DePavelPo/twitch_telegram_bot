package notification

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	twitchNotificationBGSync = "twitchNotification_BGSync"
	// tokenInvalid          = "token invalid"
)

func (tn *TwitchNotificationService) Sync(ctx context.Context) error {

	var lastId uint64 = 0

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
			users = append(users, notification.TwitchUser)
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
							logrus.Infof("cannot parse %s to uint64", streamInfo.StreamId)
							tx.Rollback()
							continue
						}

						err = tn.AddTwitchNotificationLog(ctx, tx, streamIdInt, notification.ID)
						if err != nil {
							logrus.Infof("cannot add notification log for %d", streamIdInt)
							tx.Rollback()
							continue
						}
						err = tn.ThrowNotification(ctx, streamInfo, notification.ChatId)
						if err != nil {
							tx.Rollback()
							return errors.Wrap(err, "ThrowNotification")
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
	ID         uint64 `db:"id"`
	ChatId     uint64 `db:"chat_id"`
	TwitchUser string `db:"twitch_user"`
}

const batchSize = 100

func (tn *TwitchNotificationService) GetTwitchNotificationsBatch(ctx context.Context,
	lastId uint64) (notifInfo []GetTwitchNotificationsResponse, err error) {

	query := `
				select 
					tn.id, 
					tn.chat_id, 
					tn.twitch_user 
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

func (tn *TwitchNotificationService) AddTwitchNotification(ctx context.Context, chatId uint64, user string) (err error) {

	query := `
				insert into twitch_notifications (chat_id, twitch_user) 
					values ($1, $2)
				on conflict (chat_id, twitch_user) 
					do update
					set is_active = true;
	`

	res, err := tn.db.ExecContext(ctx, query, chatId, user)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return
}

func (tn *TwitchNotificationService) SetInactiveNotification(ctx context.Context, chatId uint64, user string) (err error) {

	query := `
				update twitch_notifications 
					set is_active = false
					where (chat_id, twitch_user) = ($1, $2);
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

func (tn *TwitchNotificationService) AddTwitchNotificationLog(ctx context.Context, tx *sqlx.Tx, streamId, requestId uint64) (err error) {

	query := `
				insert into twitch_notifications_log (stream_id, request_id) 
					values ($1, $2)
				on conflict (stream_id, request_id) do nothing;
	`

	res, err := tx.ExecContext(ctx, query, streamId, requestId)
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
					tn.chat_id, 
					tn.twitch_user, 
					tnl.stream_id 
				from twitch_notifications tn 
				left join twitch_notifications_log tnl 
					on tn.id = tnl.request_id
				where tnl.stream_id = $1;
			`

	err = tn.db.GetContext(ctx, logInfo, query, streamId)
	if err != nil {
		return &GetTwitchNotificationsResponse{}, errors.Wrap(err, "GetTwitchNotificationLogByStreamId getContext")
	}

	return
}