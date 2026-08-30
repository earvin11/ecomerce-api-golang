package repository

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"ecomerce-api/internal/domain"
)

//go:embed schema.sql
var schema string

func NewPool(ctx context.Context, dsn string, maxConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("repository: parse dsn: %w", err)
	}
	cfg.MaxConns = maxConns

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("repository: create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("repository: ping database: %w", err)
	}

	return pool, nil
}

func EnsureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("repository: ensure schema: %w", err)
	}
	return nil
}

func SeedRoles(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `INSERT INTO roles (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`
	for name := range domain.AllowedRoles {
		if _, err := pool.Exec(ctx, q, name); err != nil {
			return fmt.Errorf("repository: seed role %s: %w", name, err)
		}
	}
	return nil
}

func SeedAdminUser(ctx context.Context, pool *pgxpool.Pool, email, password, username string) error {
	var roleID int
	err := pool.QueryRow(ctx, `SELECT id FROM roles WHERE name = 'ADMIN'`).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("repository: seed admin role lookup: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("repository: seed admin hash: %w", err)
	}
	const q = `INSERT INTO users (email, password, username, role_id)
	           VALUES ($1, $2, $3, $4) ON CONFLICT (username) DO NOTHING`
	if _, err := pool.Exec(ctx, q, email, string(hash), username, roleID); err != nil {
		return fmt.Errorf("repository: seed admin user: %w", err)
	}
	return nil
}
