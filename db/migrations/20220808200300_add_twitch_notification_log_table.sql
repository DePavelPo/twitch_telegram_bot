-- +goose Up
-- +goose StatementBegin
create table twitch_notifications_log
(
    id bigserial not null primary key,
    stream_id bigint not null,
    chat_id bigint not null,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null
);

create unique index stream_id_chat_id_unique_key on twitch_notifications_log(stream_id, chat_id);

comment on column twitch_notifications_log.chat_id is 'telegram chat ID';
comment on column twitch_notifications_log.stream_id is 'twitch stream ID';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists stream_id_chat_id_unique_key;
drop table if exists twitch_notifications_log;
-- +goose StatementEnd