package postgres

import (
	"context"
	"database/sql"
	"errors"

	"cdek_wishlist/internal/domain"
)

type WishlistRepository struct {
	db *sql.DB
}

func NewWishlistRepository(db *sql.DB) *WishlistRepository { return &WishlistRepository{db: db} }

func (r *WishlistRepository) Create(ctx context.Context, wishlist *domain.Wishlist) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO wishlists (user_id, name, description, event_date, token, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		wishlist.UserID, wishlist.Name, wishlist.Description, wishlist.EventDate, wishlist.Token, wishlist.CreatedAt,
	).Scan(&wishlist.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (r *WishlistRepository) Update(ctx context.Context, wishlist *domain.Wishlist) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE wishlists SET name = $1, description = $2, event_date = $3 WHERE id = $4`,
		wishlist.Name, wishlist.Description, wishlist.EventDate, wishlist.ID,
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

func (r *WishlistRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM wishlists WHERE id = $1`, id)
	return err
}

func (r *WishlistRepository) FindByID(ctx context.Context, id int64) (*domain.Wishlist, error) {
	var wishlist domain.Wishlist
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, description, event_date, token, created_at FROM wishlists WHERE id = $1`, id,
	).Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.EventDate, &wishlist.Token, &wishlist.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &wishlist, nil
}

func (r *WishlistRepository) FindByToken(ctx context.Context, token string) (*domain.Wishlist, error) {
	var wishlist domain.Wishlist
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, description, event_date, token, created_at FROM wishlists WHERE token = $1`, token,
	).Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.EventDate, &wishlist.Token, &wishlist.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &wishlist, nil
}

func (r *WishlistRepository) ListByUserID(ctx context.Context, userID int64) ([]domain.Wishlist, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, description, event_date, token, created_at FROM wishlists WHERE user_id = $1 ORDER BY id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Wishlist, 0)
	for rows.Next() {
		var wishlist domain.Wishlist
		if err := rows.Scan(&wishlist.ID, &wishlist.UserID, &wishlist.Name, &wishlist.Description, &wishlist.EventDate, &wishlist.Token, &wishlist.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, wishlist)
	}
	return result, rows.Err()
}
