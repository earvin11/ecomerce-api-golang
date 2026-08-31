package usecases

import (
	"context"

	"ecomerce-api/internal/domain"
	"ecomerce-api/internal/repository"
)

type ProductUseCases struct {
	productRepository repository.ProductRepository
}

func NewProductUseCases(productRepository repository.ProductRepository) *ProductUseCases {
	return &ProductUseCases{productRepository: productRepository}
}

func (u *ProductUseCases) Create(ctx context.Context, data *domain.Product, actorID int) (*domain.Product, error) {
	return u.productRepository.Create(ctx, data, actorID)
}

func (u *ProductUseCases) GetAll(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	return u.productRepository.GetAll(ctx, page, pageSize)
}

func (u *ProductUseCases) GetById(ctx context.Context, id int) (*domain.Product, error) {
	return u.productRepository.GetById(ctx, id)
}

func (u *ProductUseCases) Update(ctx context.Context, id int, data *domain.UpdateProduct, actorID int) (*domain.Product, error) {
	return u.productRepository.Update(ctx, id, data, actorID)
}

func (u *ProductUseCases) Delete(ctx context.Context, id int) error {
	return u.productRepository.Delete(ctx, id)
}
