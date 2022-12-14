-- +goose Up
-- +goose StatementBegin
create table twitch_notifications_log
(
    id bigserial not null primary key,
    stream_id bigint not null,
    request_id bigint references twitch_notifications(id) not null,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null
);

create unique index stream_id_request_id_unique_key on twitch_notifications_log(stream_id, request_id);

comment on column twitch_notifications_log.stream_id is 'ID трансляции';
comment on column twitch_notifications_log.request_id is 'ID запроса на нотификацию';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index if exists stream_id_request_id_unique_key;
drop table if exists twitch_notifications_log;
-- +goose StatementEnd