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

type UserRepository interface {
	Create(ctx context.Context, data *domain.User) (*domain.User, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*domain.User, int, error)
	GetById(ctx context.Context, id int) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, id int, data *domain.UpdateUser) (*domain.User, error)
	Delete(ctx context.Context, id int) error
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

const selectUserWithRole = `SELECT u.id, u.email, u.password, u.username, u.role_id, u.photo, r.id, r.name
                            FROM users u JOIN roles r ON r.id = u.role_id`

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	var roleID, roleID2 int
	var roleName string
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.Username, &roleID, &u.Photo, &roleID2, &roleName); err != nil {
		return nil, err
	}
	u.RoleID = roleID
	u.Role = &domain.Role{ID: roleID2, Name: roleName}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, data *domain.User) (*domain.User, error) {
	const q = `INSERT INTO users (email, password, username, role_id, photo)
	           VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id int
	if err := r.db.QueryRow(ctx, q, data.Email, data.Password, data.Username, data.RoleID, data.Photo).Scan(&id); err != nil {
		return nil, handleUserConstraint(err, "create")
	}
	return r.GetById(ctx, id)
}

func (r *userRepository) GetAll(ctx context.Context, page, pageSize int) ([]*domain.User, int, error) {
	const qCount = `SELECT COUNT(*) FROM users`
	var total int
	if err := r.db.QueryRow(ctx, qCount).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("user repository count: %w", err)
	}

	const q = `SELECT u.id, u.email, u.password, u.username, u.role_id, u.photo, r.id, r.name
	           FROM users u JOIN roles r ON r.id = u.role_id
	           ORDER BY u.id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("user repository list: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("user repository scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("user repository iterate: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) GetById(ctx context.Context, id int) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, selectUserWithRole+` WHERE u.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: user %d", domain.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("user repository get by id: %w", err)
	}
	return u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, selectUserWithRole+` WHERE u.username = $1`, username))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: user %s", domain.ErrNotFound, username)
	}
	if err != nil {
		return nil, fmt.Errorf("user repository get by username: %w", err)
	}
	return u, nil
}

func (r *userRepository) Update(ctx context.Context, id int, data *domain.UpdateUser) (*domain.User, error) {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)
	idx := 1

	if data.Email != nil {
		sets = append(sets, fmt.Sprintf("email = $%d", idx))
		args = append(args, *data.Email)
		idx++
	}
	if data.Password != nil {
		sets = append(sets, fmt.Sprintf("password = $%d", idx))
		args = append(args, *data.Password)
		idx++
	}
	if data.Username != nil {
		sets = append(sets, fmt.Sprintf("username = $%d", idx))
		args = append(args, *data.Username)
		idx++
	}
	if data.RoleID != nil {
		sets = append(sets, fmt.Sprintf("role_id = $%d", idx))
		args = append(args, *data.RoleID)
		idx++
	}
	if data.Photo.Set {
		sets = append(sets, fmt.Sprintf("photo = $%d", idx))
		args = append(args, data.Photo.Value)
		idx++
	}

	if len(sets) == 0 {
		return r.GetById(ctx, id)
	}

	var rowID int
	q := fmt.Sprintf(`UPDATE users SET %s WHERE id = $%d RETURNING id`, strings.Join(sets, ", "), idx)
	args = append(args, id)
	err := r.db.QueryRow(ctx, q, args...).Scan(&rowID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: user %d", domain.ErrNotFound, id)
		}
		return nil, handleUserConstraint(err, "update")
	}
	return r.GetById(ctx, rowID)
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	const q = `DELETE FROM users WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("user repository delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: user %d", domain.ErrNotFound, id)
	}
	return nil
}

func handleUserConstraint(err error, action string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			field := "email"
			if pgErr.ConstraintName == "users_username_key" {
				field = "username"
			}
			return fmt.Errorf("%w: %s already in use", domain.ErrInvalidData, field)
		case "23503":
			return fmt.Errorf("%w: referenced role does not exist", domain.ErrInvalidData)
		}
	}
	return fmt.Errorf("user repository %s: %w", action, err)
}
