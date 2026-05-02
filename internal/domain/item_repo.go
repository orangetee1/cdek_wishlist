package domain

import "context"

type ItemRepository interface {
	Create(ctx context.Context, item *Item) error
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Item, error)
	ListByWishlistID(ctx context.Context, wishlistID int64) ([]Item, error)
	Book(ctx context.Context, id int64) error
}
