package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}
