package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecomerce-api/internal/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, data *domain.Category, actorID int) (*domain.Category, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*domain.Category, int, error)
	GetById(ctx context.Context, id int) (*domain.Category, error)
	Update(ctx context.Context, id int, data *domain.UpdateCategory, actorID int) (*domain.Category, error)
	Delete(ctx context.Context, id int) error
}

const selectCategory = `SELECT id, name, created_by, created_at, updated_by, updated_at FROM categories`

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{db: db}
}

func scanCategory(row pgx.Row) (*domain.Category, error) {
	var c domain.Category
	if err := row.Scan(&c.ID, &c.Name, &c.CreatedBy, &c.CreatedAt, &c.UpdatedBy, &c.UpdatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *categoryRepository) Create(ctx context.Context, data *domain.Category, actorID int) (*domain.Category, error) {
	const q = `INSERT INTO categories (name, created_by, created_at, updated_by, updated_at)
	           VALUES ($1, $2, now(), $3, now())
	           RETURNING id, name, created_by, created_at, updated_by, updated_at`
	c, err := scanCategory(r.db.QueryRow(ctx, q, data.Name, actorID, actorID))
	if err != nil {
		return nil, fmt.Errorf("category repository create: %w", err)
	}
	return c, nil
}

func (r *categoryRepository) GetAll(ctx context.Context, page, pageSize int) ([]*domain.Category, int, error) {
	const qCount = `SELECT COUNT(*) FROM categories`
	var total int
	if err := r.db.QueryRow(ctx, qCount).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("category repository count: %w", err)
	}

	const q = selectCategory + ` ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("category repository list: %w", err)
	}
	defer rows.Close()

	categories := make([]*domain.Category, 0)
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("category repository scan: %w", err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("category repository iterate: %w", err)
	}

	return categories, total, nil
}

func (r *categoryRepository) GetById(ctx context.Context, id int) (*domain.Category, error) {
	c, err := scanCategory(r.db.QueryRow(ctx, selectCategory+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: category %d", domain.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("category repository get by id: %w", err)
	}
	return c, nil
}

func (r *categoryRepository) Update(ctx context.Context, id int, data *domain.UpdateCategory, actorID int) (*domain.Category, error) {
	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)
	idx := 1

	if data.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *data.Name)
		idx++
	}

	if len(sets) == 0 {
		return r.GetById(ctx, id)
	}

	sets = append(sets, fmt.Sprintf("updated_by = $%d", idx))
	args = append(args, actorID)
	idx++
	sets = append(sets, "updated_at = now()")

	q := fmt.Sprintf(`UPDATE categories SET %s WHERE id = $%d RETURNING id, name, created_by, created_at, updated_by, updated_at`, strings.Join(sets, ", "), idx)
	args = append(args, id)

	c, err := scanCategory(r.db.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: category %d", domain.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("category repository update: %w", err)
	}
	return c, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	const q = `DELETE FROM categories WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("category repository delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: category %d", domain.ErrNotFound, id)
	}
	return nil
}
