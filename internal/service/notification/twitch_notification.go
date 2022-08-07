package twitch_notification

import (
	"context"
	twitch_client "twitch_telegram_bot/internal/client/twitch-client"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type TwitchNotificationService struct {
	db           *sqlx.DB
	twitchClient *twitch_client.TwitchClient
}

func NewTwitchNotificationService(db *sqlx.DB, twitchClient *twitch_client.TwitchClient) (*TwitchNotificationService, error) {
	service := &TwitchNotificationService{
		db:           db,
		twitchClient: twitchClient,
	}

	ctx := context.Background()
	err := service.Sync(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Sync")
	}

	return service, nil
}

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

						err := tn.ThrowNotification(ctx, streamInfo, notification.ChatId)
						if err != nil {
							return errors.Wrap(err, "ThrowNotification")
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

	err = tn.db.SelectContext(ctx, notifInfo, query, lastId, batchSize)
	if err != nil {
		return []GetTwitchNotificationsResponse{}, errors.Wrap(err, "GetTwitchNotificationsBatch selectContext")
	}

	return
}

func (tn *TwitchNotificationService) AddTwitchNotification(ctx context.Context, tx *sqlx.Tx, chatId uint64, user string) (err error) {

	query := `
				insert into twitch_notifications (chat_id, twitch_user) 
					values ($1, $2)
				on conflict (chat_id, twitch_user) 
					do update
					set is_active = true;
	`

	res, err := tx.ExecContext(ctx, query, chatId, user)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return
}

func (tn *TwitchNotificationService) SetInactiveNotification(ctx context.Context, tx *sqlx.Tx, chatId uint64, user string) (err error) {

	query := `
				update twitch_notifications 
					set is_active = false
					where (chat_id, twitch_user) = ($1, $2);
	`

	res, err := tx.ExecContext(ctx, query, chatId, user)
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
