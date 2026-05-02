Wishlist API
============

REST API для регистрации пользователей, управления вишлистами и бронирования подарков по публичной ссылке.

Run with Docker Compose:

```bash
docker compose up --build
```

Migrations are applied automatically on startup with `goose`.

Environment (see .env.example):
- `DATABASE_DSN` - Postgres DSN
- `JWT_SECRET` - JWT signing secret
- `PORT` - app port
- `MIGRATIONS_DIR` - goose migrations directory

Endpoints:
- `POST /auth/register`
- `POST /auth/login`
- Authenticated (Header: `Authorization: Bearer <token>`):
  - `GET /api/wishlists`
  - `POST /api/wishlists`
  - `GET /api/wishlists/:id`
  - `PUT /api/wishlists/:id`
  - `DELETE /api/wishlists/:id`
  - `POST /api/wishlists/:id/items`
  - `PUT /api/wishlists/:id/items/:item_id`
  - `DELETE /api/wishlists/:id/items/:item_id`
- Public:
  - `GET /public/wishlists/:token`
  - `POST /public/wishlists/:token/items/:item_id/book`

Example login:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"secret123"}'
```
