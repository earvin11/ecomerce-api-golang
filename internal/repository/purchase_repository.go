package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"ecomerce-api/internal/domain"
)

type PurchaseRepository interface {
	Create(ctx context.Context, purchase *domain.Purchase) (*domain.Purchase, error)
	GetAllByUser(ctx context.Context, userID int, page, pageSize int) ([]*domain.Purchase, int, error)
}

type purchaseRepository struct {
	db *pgxpool.Pool
}

func NewPurchaseRepository(db *pgxpool.Pool) PurchaseRepository {
	return &purchaseRepository{db: db}
}

func (r *purchaseRepository) Create(ctx context.Context, p *domain.Purchase) (*domain.Purchase, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("purchase repository begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var quantity int
	var inStock bool
	if err := tx.QueryRow(ctx, `SELECT quantity, in_stock, name FROM products WHERE id = $1 FOR UPDATE`, p.ProductID).
		Scan(&quantity, &inStock, &p.ProductName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: product %d", domain.ErrNotFound, p.ProductID)
		}
		return nil, fmt.Errorf("purchase repository lock product: %w", err)
	}
	if !inStock || quantity < p.Quantity {
		return nil, fmt.Errorf("%w: insufficient stock for product %d", domain.ErrInvalidData, p.ProductID)
	}

	if _, err := tx.Exec(ctx, `UPDATE products SET quantity = quantity - $1, updated_by = $2, updated_at = now() WHERE id = $3`,
		p.Quantity, p.UserID, p.ProductID); err != nil {
		return nil, fmt.Errorf("purchase repository decrement stock: %w", err)
	}

	const q = `INSERT INTO purchases (user_id, product_id, quantity, unit_price, discount, total, wallet_ref)
	           VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`
	if err := tx.QueryRow(ctx, q, p.UserID, p.ProductID, p.Quantity, int64(p.UnitPrice), int64(p.Discount), int64(p.Total), p.WalletRef).
		Scan(&p.ID, &p.CreatedAt); err != nil {
		return nil, fmt.Errorf("purchase repository insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("purchase repository commit: %w", err)
	}

	return p, nil
}

func (r *purchaseRepository) GetAllByUser(ctx context.Context, userID int, page, pageSize int) ([]*domain.Purchase, int, error) {
	const qCount = `SELECT COUNT(*) FROM purchases WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, qCount, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("purchase repository count: %w", err)
	}

	const q = `SELECT pu.id, pu.user_id, pu.product_id, pr.name, pu.quantity, pu.unit_price, pu.discount, pu.total, pu.wallet_ref, pu.created_at
	           FROM purchases pu JOIN products pr ON pr.id = pu.product_id
	           WHERE pu.user_id = $1 ORDER BY pu.id DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("purchase repository list: %w", err)
	}
	defer rows.Close()

	purchases := make([]*domain.Purchase, 0)
	for rows.Next() {
		var p domain.Purchase
		var unitPrice, discount, total int64
		if err := rows.Scan(&p.ID, &p.UserID, &p.ProductID, &p.ProductName, &p.Quantity, &unitPrice, &discount, &total, &p.WalletRef, &p.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("purchase repository scan: %w", err)
		}
		p.UnitPrice = domain.Cents(unitPrice)
		p.Discount = domain.Cents(discount)
		p.Total = domain.Cents(total)
		purchases = append(purchases, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("purchase repository iterate: %w", err)
	}

	return purchases, total, nil
}
