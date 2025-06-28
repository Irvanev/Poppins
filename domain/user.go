package domain

import "time"

type User struct {
	ID               int64     `json:"id"`
	TelegramID       int64     `json:"telegram_id"`
	Name             string    `json:"name"`
	Phone            string    `json:"phone"`
	PreferredContact string    `json:"preferred_contact"`
	AdsCount         int64     `json:"ads_count"`
	CreatedAt        time.Time `json:"created_at"`
}
