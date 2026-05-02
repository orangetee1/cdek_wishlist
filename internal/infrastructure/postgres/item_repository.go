package postgres

import (
	"context"
	"database/sql"
	"errors"

	"cdek_wishlist/internal/domain"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository { return &ItemRepository{db: db} }

func (r *ItemRepository) Create(ctx context.Context, item *domain.Item) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO items (wishlist_id, name, description, link, desire_degree, booked, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, now()) RETURNING id`,
		item.WishlistID, item.Name, item.Description, item.Link, int(item.DesireDegree), item.Booked,
	).Scan(&item.ID)
}

func (r *ItemRepository) Update(ctx context.Context, item *domain.Item) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET name = $1, description = $2, link = $3, desire_degree = $4, booked = $5 WHERE id = $6`,
		item.Name, item.Description, item.Link, int(item.DesireDegree), item.Booked, item.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ItemRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM items WHERE id = $1`, id)
	return err
}

func (r *ItemRepository) FindByID(ctx context.Context, id int64) (*domain.Item, error) {
	var item domain.Item
	var degree int
	err := r.db.QueryRowContext(ctx,
		`SELECT id, wishlist_id, name, description, link, desire_degree, booked FROM items WHERE id = $1`, id,
	).Scan(&item.ID, &item.WishlistID, &item.Name, &item.Description, &item.Link, &degree, &item.Booked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	item.DesireDegree = domain.Degree(degree)
	return &item, nil
}

func (r *ItemRepository) ListByWishlistID(ctx context.Context, wishlistID int64) ([]domain.Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, wishlist_id, name, description, link, desire_degree, booked FROM items WHERE wishlist_id = $1 ORDER BY id ASC`,
		wishlistID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		var degree int
		if err := rows.Scan(&item.ID, &item.WishlistID, &item.Name, &item.Description, &item.Link, &degree, &item.Booked); err != nil {
			return nil, err
		}
		item.DesireDegree = domain.Degree(degree)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *ItemRepository) Book(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE items SET booked = TRUE WHERE id = $1 AND booked = FALSE`, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAlreadyBooked
	}
	return nil
}
