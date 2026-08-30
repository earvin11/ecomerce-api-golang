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

type RoleRepository interface {
	Create(ctx context.Context, data *domain.Role) (*domain.Role, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*domain.Role, int, error)
	GetById(ctx context.Context, id int) (*domain.Role, error)
	Update(ctx context.Context, id int, data *domain.UpdateRole) (*domain.Role, error)
	Delete(ctx context.Context, id int) error
}

type roleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, data *domain.Role) (*domain.Role, error) {
	const q = `INSERT INTO roles (name) VALUES ($1) RETURNING id, name`
	row := r.db.QueryRow(ctx, q, data.Name)
	if err := row.Scan(&data.ID, &data.Name); err != nil {
		return nil, handleRoleConstraint(err, "create")
	}
	return data, nil
}

func (r *roleRepository) GetAll(ctx context.Context, page, pageSize int) ([]*domain.Role, int, error) {
	const qCount = `SELECT COUNT(*) FROM roles`
	var total int
	if err := r.db.QueryRow(ctx, qCount).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("role repository count: %w", err)
	}

	const q = `SELECT id, name FROM roles ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("role repository list: %w", err)
	}
	defer rows.Close()

	roles := make([]*domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, 0, fmt.Errorf("role repository scan: %w", err)
		}
		roles = append(roles, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("role repository iterate: %w", err)
	}

	return roles, total, nil
}

func (r *roleRepository) GetById(ctx context.Context, id int) (*domain.Role, error) {
	const q = `SELECT id, name FROM roles WHERE id = $1`
	var role domain.Role
	err := r.db.QueryRow(ctx, q, id).Scan(&role.ID, &role.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: role %d", domain.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("role repository get by id: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, id int, data *domain.UpdateRole) (*domain.Role, error) {
	sets := make([]string, 0, 1)
	args := make([]any, 0, 2)
	idx := 1

	if data.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", idx))
		args = append(args, *data.Name)
		idx++
	}

	if len(sets) == 0 {
		return r.GetById(ctx, id)
	}

	q := fmt.Sprintf(`UPDATE roles SET %s WHERE id = $%d RETURNING id, name`, strings.Join(sets, ", "), idx)
	args = append(args, id)

	var role domain.Role
	err := r.db.QueryRow(ctx, q, args...).Scan(&role.ID, &role.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: role %d", domain.ErrNotFound, id)
		}
		return nil, handleRoleConstraint(err, "update")
	}
	return &role, nil
}

func (r *roleRepository) Delete(ctx context.Context, id int) error {
	const q = `DELETE FROM roles WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("role repository delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: role %d", domain.ErrNotFound, id)
	}
	return nil
}

func handleRoleConstraint(err error, action string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: role name already in use", domain.ErrInvalidData)
	}
	return fmt.Errorf("role repository %s: %w", action, err)
}
