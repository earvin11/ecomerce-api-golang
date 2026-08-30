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
	Create(ctx context.Context, data *domain.Product) (*domain.Product, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error)
	GetById(ctx context.Context, id int) (*domain.Product, error)
	Update(ctx context.Context, id int, data *domain.UpdateProduct) (*domain.Product, error)
	Delete(ctx context.Context, id int) error
}

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, data *domain.Product) (*domain.Product, error) {
	const q = `INSERT INTO products (name, price, category, in_stock, quantity, img)
	           VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, name, price, category, in_stock, quantity, img`
	var price int64
	row := r.db.QueryRow(ctx, q, data.Name, int64(data.Price), data.Category, data.InStock, data.Quantity, data.Img)
	if err := row.Scan(&data.ID, &data.Name, &price, &data.Category, &data.InStock, &data.Quantity, &data.Img); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%w: referenced category does not exist", domain.ErrInvalidData)
		}
		return nil, fmt.Errorf("product repository create: %w", err)
	}
	data.Price = domain.Cents(price)
	return data, nil
}

func (r *productRepository) GetAll(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	const qCount = `SELECT COUNT(*) FROM products`
	var total int
	if err := r.db.QueryRow(ctx, qCount).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("product repository count: %w", err)
	}

	const q = `SELECT id, name, price, category, in_stock, quantity, img
	           FROM products ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("product repository list: %w", err)
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		var p domain.Product
		var price int64
		if err := rows.Scan(&p.ID, &p.Name, &price, &p.Category, &p.InStock, &p.Quantity, &p.Img); err != nil {
			return nil, 0, fmt.Errorf("product repository scan: %w", err)
		}
		p.Price = domain.Cents(price)
		products = append(products, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("product repository iterate: %w", err)
	}

	return products, total, nil
}

func (r *productRepository) GetById(ctx context.Context, id int) (*domain.Product, error) {
	const q = `SELECT id, name, price, category, in_stock, quantity, img FROM products WHERE id = $1`
	var p domain.Product
	var price int64
	err := r.db.QueryRow(ctx, q, id).Scan(&p.ID, &p.Name, &price, &p.Category, &p.InStock, &p.Quantity, &p.Img)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: product %d", domain.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("product repository get by id: %w", err)
	}
	p.Price = domain.Cents(price)
	return &p, nil
}

func (r *productRepository) Update(ctx context.Context, id int, data *domain.UpdateProduct) (*domain.Product, error) {
	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)
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

	q := fmt.Sprintf(`UPDATE products SET %s WHERE id = $%d
	                  RETURNING id, name, price, category, in_stock, quantity, img`, strings.Join(sets, ", "), idx)
	args = append(args, id)

	var p domain.Product
	var price int64
	err := r.db.QueryRow(ctx, q, args...).Scan(&p.ID, &p.Name, &price, &p.Category, &p.InStock, &p.Quantity, &p.Img)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, fmt.Errorf("%w: referenced category does not exist", domain.ErrInvalidData)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: product %d", domain.ErrNotFound, id)
		}
		return nil, fmt.Errorf("product repository update: %w", err)
	}
	p.Price = domain.Cents(price)
	return &p, nil
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
