-- +goose Up
-- +goose StatementBegin
create table twitch_notifications
(
    id bigserial not null primary key,
    chat_id bigint not null,
    twitch_user text not null,
    is_active boolean default true not null,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null
);

create unique index chat_id_twitch_user_unique_key on twitch_notifications(chat_id, twitch_user);

comment on column twitch_notifications.chat_id is 'ID телеграм чата';
comment on column twitch_notifications.twitch_user is 'наименование пользователя твича, по которому нужны нотификации';
comment on column twitch_notifications.is_active is 'активен ли запрос на нотификацию';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists chat_id_twitch_user_unique_key;
drop table if exists twitch_notifications;
-- +goose StatementEnd