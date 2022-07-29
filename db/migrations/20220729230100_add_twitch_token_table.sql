-- +goose Up
-- +goose StatementBegin
create table twitch_tokens
(
    id bigserial not null primary key,
    token text not null,
    is_expired boolean default false not null,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists twitch_tokens;
-- +goose StatementEnd