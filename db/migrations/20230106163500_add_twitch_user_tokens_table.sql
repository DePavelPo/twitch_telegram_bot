-- +goose Up
-- +goose StatementBegin

create table twitch_user_tokens
(
    id bigserial not null primary key,
    chat_id bigint not null unique,
    access_token text unique,
    refresh_token text unique,
    scope text[] default array[]::text[] not null,
    current_state text not null unique,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null
);

comment on column twitch_user_tokens.scope is 'list of twitch scopes';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists twitch_user_tokens;
-- +goose StatementEnd