package models

import "time"

type GetUserInfoResponse struct {
	Data []TwitchUserInfo `json:"data"`
}

type TwitchUserInfo struct {
	UserID          string    `json:"id"`                // User’s ID
	Login           string    `json:"login"`             // User’s login name
	DisplayName     string    `json:"display_name"`      // User’s display name
	Type            string    `json:"type"`              // User’s type: "staff", "admin", "global_mod", or ""
	BroadcasterType string    `json:"broadcaster_type"`  // User’s broadcaster type: "partner", "affiliate", or ""
	Description     string    `json:"description"`       // User’s channel description
	ProfileImageUrl string    `json:"profile_image_url"` // URL of the user’s profile image
	OfflineImageUrl string    `json:"offline_image_url"` // URL of the user’s offline image
	ViewCount       uint64    `json:"view_count"`        // Total number of views of the user’s channel. DEPRECATED!!! Last update 15.04.2022
	CreatedAt       time.Time `json:"created_at"`        // Date when the user was created
}
