package domain

import "time"

type Wishlist struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	EventDate   time.Time
	Token       string
	CreatedAt   time.Time
	Items       []Item
}
