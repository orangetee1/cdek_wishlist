package usecase

import "time"

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type WishlistInput struct {
	Name        string
	Description string
	EventDate   time.Time
}

type ItemInput struct {
	Name        string
	Description string
	Link        string
	Degree      int
}
