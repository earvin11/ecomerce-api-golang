package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecomerce-api/internal/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, data *domain.Product, actorID int) (*domain.Product, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error)
	GetById(ctx context.Context, id int) (*domain.Product, error)
	Update(ctx context.Context, id int, data *domain.UpdateProduct, actorID int) (*domain.Product, error)
	Delete(ctx context.Context, id int) error
}

const selectProduct = `SELECT id, name, price, category, in_stock, quantity, img, created_by, created_at, updated_by, updated_at
                       FROM products`

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var p domain.Product
	var price int64
	if err := row.Scan(&p.ID, &p.Name, &price, &p.Category, &p.InStock, &p.Quantity, &p.Img, &p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Price = domain.Cents(price)
	return &p, nil
}

func handleProductConstraint(err error, action string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return fmt.Errorf("%w: referenced category does not exist", domain.ErrInvalidData)
	}
	return fmt.Errorf("product repository %s: %w", action, err)
}

func (r *productRepository) Create(ctx context.Context, data *domain.Product, actorID int) (*domain.Product, error) {
	const q = `INSERT INTO products (name, price, category, in_stock, quantity, img, created_by, created_at, updated_by, updated_at)
	           VALUES ($1, $2, $3, $4, $5, $6, $7, now(), $8, now())
	           RETURNING id, name, price, category, in_stock, quantity, img, created_by, created_at, updated_by, updated_at`
	p, err := scanProduct(r.db.QueryRow(ctx, q, data.Name, int64(data.Price), data.Category, data.InStock, data.Quantity, data.Img, actorID, actorID))
	if err != nil {
		return nil, handleProductConstraint(err, "create")
	}
	return p, nil
}

func (r *productRepository) GetAll(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	const qCount = `SELECT COUNT(*) FROM products`
	var total int
	if err := r.db.QueryRow(ctx, qCount).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("product repository count: %w", err)
	}

	const q = selectProduct + ` ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("product repository list: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("product repository scan: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("product repository iterate: %w", err)
	}

	return products, total, nil
}

func (r *productRepository) GetById(ctx context.Context, id int) (*domain.Product, error) {
	p, err := scanProduct(r.db.QueryRow(ctx, selectProduct+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: product %d", domain.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("product repository get by id: %w", err)
	}
	return p, nil
}

func (r *productRepository) Update(ctx context.Context, id int, data *domain.UpdateProduct, actorID int) (*domain.Product, error) {
	sets := make([]string, 0, 6)
	args := make([]any, 0, 7)
	idx := 1

	if data.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *data.Name)
		idx++
	}
	if data.Price != nil {
		sets = append(sets, fmt.Sprintf("price = $%d", idx))
		args = append(args, int64(*data.Price))
		idx++
	}
	if data.Category != nil {
		sets = append(sets, fmt.Sprintf("category = $%d", idx))
		args = append(args, *data.Category)
		idx++
	}
	if data.InStock != nil {
		sets = append(sets, fmt.Sprintf("in_stock = $%d", idx))
		args = append(args, *data.InStock)
		idx++
	}
	if data.Quantity != nil {
		sets = append(sets, fmt.Sprintf("quantity = $%d", idx))
		args = append(args, *data.Quantity)
		idx++
	}
	if data.Img.Set {
		sets = append(sets, fmt.Sprintf("img = $%d", idx))
		args = append(args, data.Img.Value)
		idx++
	}

	if len(sets) == 0 {
		return r.GetById(ctx, id)
	}

	sets = append(sets, fmt.Sprintf("updated_by = $%d", idx))
	args = append(args, actorID)
	idx++
	sets = append(sets, "updated_at = now()")

	q := fmt.Sprintf(`UPDATE products SET %s WHERE id = $%d
	                  RETURNING id, name, price, category, in_stock, quantity, img, created_by, created_at, updated_by, updated_at`, strings.Join(sets, ", "), idx)
	args = append(args, id)

	p, err := scanProduct(r.db.QueryRow(ctx, q, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: product %d", domain.ErrNotFound, id)
		}
		return nil, handleProductConstraint(err, "update")
	}
	return p, nil
}

func (r *productRepository) Delete(ctx context.Context, id int) error {
	const q = `DELETE FROM products WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("product repository delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: product %d", domain.ErrNotFound, id)
	}
	return nil
}
