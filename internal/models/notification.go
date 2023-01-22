package models

type GetTwitchNotificationsResponse struct {
	ID          uint64                 `db:"id"`
	ChatId      uint64                 `db:"chat_id"`
	TwitchUser  string                 `db:"twitch_user"`
	RequestType StreamNotificationType `db:"request_type"`
}

const NotifyRequestBatchSize = 100
