package repository

import (
	"context"
	"errors"
	"fmt"
	"twitch_telegram_bot/internal/models"

	"github.com/jmoiron/sqlx"
)

func (dbr *DBRepository) GetTwitchNotificationsBatch(ctx context.Context,
	lastId uint64) (notifInfo []models.GetTwitchNotificationsResponse, err error) {

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

	err = dbr.db.SelectContext(ctx, &notifInfo, query, lastId, models.NotifyRequestBatchSize)
	if err != nil {
		return []models.GetTwitchNotificationsResponse{}, err
	}

	return
}

func (dbr *DBRepository) AddTwitchNotification(ctx context.Context, chatId uint64, user string, notiType models.StreamNotificationType) (err error) {

	query := `
				insert into twitch_notifications (chat_id, twitch_user, request_type) 
					values ($1, $2, $3)
				on conflict (chat_id, twitch_user, request_type) 
					do update
					set (request_type, is_active, updated_at) = ($3, true, now());
	`

	res, err := dbr.db.ExecContext(ctx, query, chatId, user, notiType)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return
}

func (dbr *DBRepository) SetInactiveNotificationByType(
	ctx context.Context,
	chatId uint64,
	user string,
	requestType models.StreamNotificationType,
) (err error) {

	var query string

	switch requestType {
	case models.NotificationByUser:

		query = fmt.Sprintf(`
				update twitch_notifications 
					set (is_active, updated_at) = (false, now())
					where (chat_id, twitch_user, request_type) = (%d, '%s', '%s');
		`, chatId, user, requestType)

	case models.NotificationFollowed:

		query = fmt.Sprintf(`
				update twitch_notifications 
					set (is_active, updated_at) = (false, now())
					where (chat_id, request_type) = (%d, '%s');
		`, chatId, requestType)

	}

	res, err := dbr.db.ExecContext(ctx, query)
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

func (dbr *DBRepository) AddTwitchNotificationLog(ctx context.Context, tx *sqlx.Tx, streamId, chatId uint64) (err error) {

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
