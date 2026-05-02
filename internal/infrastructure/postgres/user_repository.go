package postgres

import (
	"context"
	"database/sql"
	"errors"

	"cdek_wishlist/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, created_at) VALUES ($1, $2, $3) RETURNING id`,
		user.Email, user.PasswordHash, user.CreatedAt,
	).Scan(&user.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
