package models

import "time"

// URL is the database model for a shortened URL mapping.
type URL struct {
	ID          int64     `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CustomAlias bool      `json:"custom_alias"`
	CreatedAt   time.Time `json:"created_at"`
	ClickCount  int64     `json:"click_count"`
}

// ShortenRequest is the incoming payload for POST /api/v1/shorten.
type ShortenRequest struct {
	URL   string `json:"url" binding:"required"`
	Alias string `json:"alias,omitempty"`
}
