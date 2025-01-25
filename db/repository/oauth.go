package repository

import (
	"context"
	"twitch_telegram_bot/internal/models"

	"github.com/lib/pq"
)

func (dbr *DBRepository) GetTokensByChat(ctx context.Context, chatID uint64) (data models.TokenWithState, err error) {

	query := `
		select 
			access_token, 
			refresh_token, 
			current_state 
		from twitch_user_tokens tut 
		where chat_id = $1;
	`

	err = dbr.db.GetContext(ctx, &data, query, chatID)

	return
}

func (dbr *DBRepository) AddChatInfo(ctx context.Context, chatID uint64, state string) (err error) {

	query := `
		insert into twitch_user_tokens (chat_id, current_state) 
			values ($1, $2)
		on conflict do nothing;
	`
	_, err = dbr.db.ExecContext(ctx, query, chatID, state)

	return
}

func (dbr *DBRepository) UpdateChatTokens(ctx context.Context, chatID uint64, accessToken string, refreshToken string) (err error) {

	query := `
		update twitch_user_tokens 
			set (access_token, refresh_token) = ($1, $2)
		where chat_id = $3;
	`
	_, err = dbr.db.ExecContext(ctx, query, accessToken, refreshToken, chatID)
	if err != nil {
		return err
	}

	return
}

func (dbr *DBRepository) UpdateScopeAndGetTokens(ctx context.Context, scope []string, state string) (data models.TokensWithChatID, err error) {

	query := `
		update twitch_user_tokens
			set scope = $1
			where current_state = $2
			returning 
				chat_id,
				access_token, 
				refresh_token;
	`

	err = dbr.db.GetContext(ctx, &data, query, pq.StringArray(scope), state)

	return
}

func (dbr *DBRepository) GetTokensByChatID(ctx context.Context,
	chatID uint64, scope models.Scope) (data models.TokenWithState, err error) {

	query := `
			select 
				tut.access_token, 
				tut.refresh_token, 
				tut.current_state
			from twitch_user_tokens tut 
			where $1 = ANY(tut."scope") 
				and tut.chat_id = $2
			limit 1;
			`

	err = dbr.db.GetContext(ctx, &data, query, scope, chatID)

	return
}

func (dbr *DBRepository) UpdateChatTokensByState(ctx context.Context, state, accessToken, refreshToken string) (err error) {

	query := `
		update twitch_user_tokens 
			set (access_token, refresh_token, updated_at) = ($1, $2, now())
		where current_state = $3;
	`
	_, err = dbr.db.ExecContext(ctx, query, accessToken, refreshToken, state)
	if err != nil {
		return err
	}

	return
}
