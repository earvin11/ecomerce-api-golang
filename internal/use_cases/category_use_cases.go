package usecases

import (
	"context"

	"ecomerce-api/internal/domain"
	"ecomerce-api/internal/repository"
)

type CategoryUseCases struct {
	categoryRepository repository.CategoryRepository
}

func NewCategoryUseCases(categoryRepository repository.CategoryRepository) *CategoryUseCases {
	return &CategoryUseCases{categoryRepository: categoryRepository}
}

func (u *CategoryUseCases) Create(ctx context.Context, data *domain.Category, actorID int) (*domain.Category, error) {
	return u.categoryRepository.Create(ctx, data, actorID)
}

func (u *CategoryUseCases) GetAll(ctx context.Context, page, pageSize int) ([]*domain.Category, int, error) {
	return u.categoryRepository.GetAll(ctx, page, pageSize)
}

func (u *CategoryUseCases) GetById(ctx context.Context, id int) (*domain.Category, error) {
	return u.categoryRepository.GetById(ctx, id)
}

func (u *CategoryUseCases) Update(ctx context.Context, id int, data *domain.UpdateCategory, actorID int) (*domain.Category, error) {
	return u.categoryRepository.Update(ctx, id, data, actorID)
}

func (u *CategoryUseCases) Delete(ctx context.Context, id int) error {
	return u.categoryRepository.Delete(ctx, id)
}
