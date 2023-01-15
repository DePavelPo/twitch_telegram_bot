package models

import "time"

type StreamType string

var StreamLive StreamType = "live"

type GetActiveStreamInfoByUserReq struct {
	ID string `json:"id"`
}

type Streams struct {
	StreamInfo []Stream   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Stream struct {
	StreamId     string     `json:"id"`            // 	Stream ID
	UserId       string     `json:"user_id"`       // ID of the user who is streaming
	UserLogin    string     `json:"user_login"`    // Login of the user who is streaming
	UserName     string     `json:"user_name"`     // Display name corresponding to user_id
	GameId       string     `json:"game_id"`       // ID of the game being played on the stream
	GameName     string     `json:"game_name"`     // Name of the game being played
	StreamType   StreamType `json:"type"`          // Stream type: "live" or "" (in case of error)
	Title        string     `json:"title"`         // Stream title
	ViewerCount  uint64     `json:"viewer_count"`  // Number of viewers watching the stream at the time of the query
	StartedAt    time.Time  `json:"started_at"`    // UTC timestamp
	Lang         string     `json:"language"`      // Stream language
	ThumbnailUrl string     `json:"thumbnail_url"` // Thumbnail URL of the stream. Replace {width} and {height} with any values to get that size image
	TagIds       []string   `json:"tag_ids"`       // Shows tag IDs that apply to the stream
	Tags         []string   `json:"tags"`          // Shows tags that apply to the stream
	IsMature     bool       `json:"is_mature"`     // Contains mature content that may be inappropriate for younger audiences

}

type Pagination struct {
	Cursor string `json:"cursor"`
}

type StreamNotificationType string

var (
	NotificationByUser   StreamNotificationType = "by_user"
	NotificationFollowed StreamNotificationType = "followed"
)
