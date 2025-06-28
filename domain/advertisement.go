package domain

import "time"

type Advertisement struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	TelegramID  int64     `json:"telegram_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	PhotosUrls  string    `json:"photos_urls"`
	Address     string    `json:"address"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
