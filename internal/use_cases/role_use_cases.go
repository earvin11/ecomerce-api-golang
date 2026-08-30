package usecases

import (
	"context"

	"ecomerce-api/internal/domain"
	"ecomerce-api/internal/repository"
)

type RoleUseCases struct {
	roleRepository repository.RoleRepository
}

func NewRoleUseCases(roleRepository repository.RoleRepository) *RoleUseCases {
	return &RoleUseCases{roleRepository: roleRepository}
}

func (u *RoleUseCases) Create(ctx context.Context, data *domain.Role) (*domain.Role, error) {
	return u.roleRepository.Create(ctx, data)
}

func (u *RoleUseCases) GetAll(ctx context.Context, page, pageSize int) ([]*domain.Role, int, error) {
	return u.roleRepository.GetAll(ctx, page, pageSize)
}

func (u *RoleUseCases) GetById(ctx context.Context, id int) (*domain.Role, error) {
	return u.roleRepository.GetById(ctx, id)
}

func (u *RoleUseCases) Update(ctx context.Context, id int, data *domain.UpdateRole) (*domain.Role, error) {
	return u.roleRepository.Update(ctx, id, data)
}

func (u *RoleUseCases) Delete(ctx context.Context, id int) error {
	return u.roleRepository.Delete(ctx, id)
}
