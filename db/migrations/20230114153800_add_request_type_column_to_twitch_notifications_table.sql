-- +goose Up
-- +goose StatementBegin

create type stream_notification_type as enum ('by_user', 'followed');

alter table twitch_notifications add column if not exists request_type stream_notification_type not null;

drop index if exists chat_id_twitch_user_unique_key;

create unique index chat_id_twitch_user_request_type_unique_key on twitch_notifications(chat_id, twitch_user, request_type);

comment on column twitch_notifications.request_type is 'type of telegram notification request';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists chat_id_twitch_user_request_type_unique_key;
alter table twitch_notifications drop column if exists request_type;

create unique index chat_id_twitch_user_unique_key on twitch_notifications(chat_id, twitch_user);

drop type stream_notification_type;
-- +goose StatementEnd