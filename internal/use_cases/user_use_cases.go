package usecases

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"ecomerce-api/internal/domain"
	"ecomerce-api/internal/repository"
)

type UserUseCases struct {
	userRepository repository.UserRepository
}

func NewUserUseCases(userRepository repository.UserRepository) *UserUseCases {
	return &UserUseCases{userRepository: userRepository}
}

func (u *UserUseCases) Create(ctx context.Context, data *domain.CreateUser) (*domain.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("user use case hash password: %w", err)
	}

	user := &domain.User{
		Email:    data.Email,
		Password: string(hashed),
		Username: data.Username,
		RoleID:   data.RoleID,
	}
	return u.userRepository.Create(ctx, user)
}

func (u *UserUseCases) GetAll(ctx context.Context, page, pageSize int) ([]*domain.User, int, error) {
	return u.userRepository.GetAll(ctx, page, pageSize)
}

func (u *UserUseCases) GetById(ctx context.Context, id int) (*domain.User, error) {
	return u.userRepository.GetById(ctx, id)
}

func (u *UserUseCases) Update(ctx context.Context, id int, data *domain.UpdateUser) (*domain.User, error) {
	if data.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*data.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("user use case hash password: %w", err)
		}
		hash := string(hashed)
		data.Password = &hash
	}
	return u.userRepository.Update(ctx, id, data)
}

func (u *UserUseCases) Delete(ctx context.Context, id int) error {
	return u.userRepository.Delete(ctx, id)
}
