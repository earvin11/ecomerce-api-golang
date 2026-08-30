# AGENTS.md

Working rules and conventions for this repository. Read this before making changes.

## Project overview

E-commerce REST API in Go following clean architecture. Go 1.26.5, router is `net/http` standard library (Go 1.22+ method patterns), database is PostgreSQL accessed through `pgx/v5` (`pgxpool`). Dependencies: `github.com/jackc/pgx/v5`, `golang.org/x/crypto` (bcrypt), `github.com/golang-jwt/jwt/v5`.

Entry point: `cmd/api/main.go` (wires all dependencies). Endpoints live under `/api/v1`.

## Architecture rules (do not break)

- **Dependencies point inward.** `internal/domain` must never import external packages or other internal layers (no repository, use_cases, handlers, or libraries).
- **Depend on abstractions, not implementations.** `internal/use_cases` receives repository interfaces via constructor; it never touches the concrete struct. Interfaces `CategoryRepository` / `ProductRepository` are defined in `internal/repository` and returned by their `NewXxxRepository` constructors.
- **Dependency injection happens in `main.go` only**: config -> db pool -> repositories -> use cases -> handlers -> router. Do not instantiate repos/use cases inside handlers or use cases.
- Layer import direction: `handlers` -> `use_cases` -> `repository` -> `domain`. `handlers` may import `domain` for errors/entities; `handlers` must NOT import `repository` directly.
- Domain entities are plain structs with JSON tags. Update DTOs (`UpdateCategory`, `UpdateProduct`) use pointers to support partial `PATCH`. Nullable link fields (`Product.Img`, `User.Photo`): entities and Create DTOs use `*string` (nil -> DB NULL, JSON `null`); Update DTOs use `domain.NullableString` to distinguish absent (`Set=false`), explicit `null` (`Set=true`, `Value=nil`) and value. Repositories bind `NullableString.Value` (may be nil) and scan `*string` columns directly (pgx pointer-pointer plan).

## Directory layout

```
cmd/api/main.go        # wiring + http.Server + graceful shutdown
internal/config/       # env config (DB_*, SERVER_PORT, JWT_*, timeouts); Config.DSN()
internal/domain/       # entities (category.go, product.go, role.go, user.go, auth.go) + errors.go sentinels
internal/auth/         # JWT TokenService (HS256) - sign/parse access + refresh tokens
internal/repository/   # repo interfaces + pgx implementations + postgres.go + schema.sql
internal/use_cases/    # use cases holding repo interfaces
internal/handlers/     # HTTP handlers, router.go, auth middleware, response helpers (response.go)
Dockerfile             # multi-stage; runtime alpine, CGO_ENABLED=0
docker-compose.yml     # services: api + postgres:16-alpine (healthcheck, volume)
.env.example
```

## HTTP / API conventions

- Base path: `/api/v1`. Routes registered in `internal/handlers/router.go` using net/http method patterns, e.g. `GET /api/v1/categories/{id}`; read path params with `r.PathValue("id")`.
- **Standardized JSON body** (built in `internal/handlers/response.go`):
  - Success single: `{"message": "...", "data": {...}, "error": null}`
  - Success list: `{"message": "...", "data": [...], "meta": {...}, "error": null}`
  - Error: `{"message": "...", "data": null, "error": "detail"}`
- **`meta` (pagination, required on GetAll)** via `buildMeta(total, page, pageSize)`: `total`, `page`, `per_page`, `last_page`, `remaining`.
- Status codes:
  - `201` created, `200` get/update (PATCH returns the updated resource), `204` delete (no body).
  - `400` bad request (invalid JSON, validation, bad id, FK violation), `404` not found, `401` unauthorized, `500` internal.
- Errors are mapped centrally in `resolveError` using `errors.Is` against `domain.ErrNotFound` / `ErrInvalidData` / `ErrUnauthorized`.
- Use helpers `respondSuccess`, `respondSuccessList`, `respondError`, `respondNoContent`, `decodeJSON`, `parseID`, `parsePagination` instead of writing raw responses. Do not add dependencies for routing/JSON.

## Auth / JWT conventions

- `POST /api/v1/auth/login` (username + password, bcrypt compare) and `POST /api/v1/auth/refresh` are public. All other `/api/v1` routes (`/health` too is public) require `Authorization: Bearer <token>`, enforced by `handlers.AuthMiddleware.Require` wrapping each route in `router.go`; the middleware loads claims into the request context (`ctxUsername` / `ctxRole`).
- JWT HS256 signed with `JWT_SECRET` (required config). Access token TTL `JWT_ACCESS_TTL` (default 30m), refresh `JWT_REFRESH_TTL` (default 720h). Claims: `username` + `role` (plus `sub`/`iat`/`exp`). Token is built/parsed by `internal/auth` `TokenService`; it is the only place that knows the JWT library.
- Login returns `{access_token, refresh_token, token_type, expires_in}` (seconds). `/refresh` validates the refresh token, re-fetches the user to confirm existence and re-read the current role, then issues a fresh pair. No server-side token storage / revocation.
- Optional bootstrap: if `SEED_ADMIN_USERNAME` + `SEED_ADMIN_EMAIL` + `SEED_ADMIN_PASSWORD` are all set, `repository.SeedAdminUser` creates the first ADMIN user on startup (idempotent on `username`); needed since `/users` is protected.

## Data layer conventions

- PostgreSQL placeholders are `$1, $2, ...` (pgx). Use `INSERT ... RETURNING ...` to populate IDs.
- Money (product `price`) is stored as integer minor units (cents) in a `BIGINT` column. The `domain.Cents` type converts integer cents <-> decimal JSON (`10.84` <-> `1084`) via `MarshalJSON`/`UnmarshalJSON`; repositories bind/scan `int64` explicitly.
- `GetAll` returns `(items, total, error)` using `COUNT(*)` + `LIMIT/OFFSET`.
- Update supports partial `PATCH`: build dynamic `SET` clauses only from non-nil pointer fields; return the updated row via `RETURNING`.
- Map `pgx.ErrNoRows` to `domain.ErrNotFound`; map FK violation `pgconn.PgError` code `23503` to `domain.ErrInvalidData`.
- Schema is idempotent (`CREATE TABLE IF NOT EXISTS`) in `internal/repository/schema.sql`, embedded with `go:embed` and applied on startup by `EnsureSchema`.

## Config

All configuration comes from environment variables defined in `internal/config/config.go` (see `.env.example`). Defaults: `SERVER_PORT=8080`, `DB_HOST=localhost`, `DB_PORT=5432`, `DB_USER=postgres`, `DB_PASSWORD=postgres`, `DB_NAME=ecommerce`, `JWT_ACCESS_TTL=30m`, `JWT_REFRESH_TTL=720h`. `JWT_SECRET` is required (no default).

## Commands

- Build: `go build ./...`
- Vet: `go vet ./...`
- Test: `go test ./...`
- Format: `gofmt -l .` (must be empty before finishing)
- Run locally: needs a PostgreSQL DB (e.g. `docker compose up db`) then `go run ./cmd/api`
- Full stack: `docker compose up --build`

## Rules for changes

- Do not add comments to code unless asked.
- Keep responses/errors in the existing Spanish-style message model; messages are contextual per method (e.g. "Category created successfully").
- When adding an entity: add it to `domain`, create its repository interface + implementation, a use case, a handler, routes in `router.go`, and a table in `schema.sql` — and complete the wiring in `cmd/api/main.go`.
- Verify with build + vet + test before finishing.