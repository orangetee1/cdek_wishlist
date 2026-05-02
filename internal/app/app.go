package app

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cdek_wishlist/internal/config"
	"cdek_wishlist/internal/infrastructure/migration"
	"cdek_wishlist/internal/infrastructure/postgres"
	"cdek_wishlist/internal/transport/httpserver"
	"cdek_wishlist/internal/usecase"

	_ "github.com/lib/pq"
)

func Run(cfg config.Config) error {
	db, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := waitForDatabase(db, 30, time.Second); err != nil {
		return err
	}

	if err := migration.Up(db, cfg.MigrationDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	deps := usecase.Dependencies{
		Users:     postgres.NewUserRepository(db),
		Wishlists: postgres.NewWishlistRepository(db),
		Items:     postgres.NewItemRepository(db),
		Hasher:    usecase.BcryptHasher{},
		Tokens:    usecase.NewJWTManager(cfg.JWTSecret, 24*time.Hour),
		Now:       time.Now,
	}

	service := usecase.NewService(deps)
	server := httpserver.New(service, deps.Tokens)
	return server.Run(":" + cfg.Port)
}

func waitForDatabase(db *sql.DB, attempts int, delay time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(attempts)*delay)
	defer cancel()

	for i := 0; i < attempts; i++ {
		if err := db.PingContext(ctx); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("database is not reachable")
}
