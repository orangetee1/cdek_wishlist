package domain

import "context"

type WishlistRepository interface {
	Create(ctx context.Context, wishlist *Wishlist) error
	Update(ctx context.Context, wishlist *Wishlist) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Wishlist, error)
	FindByToken(ctx context.Context, token string) (*Wishlist, error)
	ListByUserID(ctx context.Context, userID int64) ([]Wishlist, error)
}
