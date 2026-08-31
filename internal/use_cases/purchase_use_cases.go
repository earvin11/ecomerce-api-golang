package usecases

import (
	"context"
	"fmt"
	"time"

	"ecomerce-api/internal/domain"
	"ecomerce-api/internal/repository"
	"ecomerce-api/internal/wallet"
)

type PurchaseUseCases struct {
	purchaseRepository repository.PurchaseRepository
	productRepository  repository.ProductRepository
	walletClient       wallet.Client
	currency           string
}

func NewPurchaseUseCases(purchaseRepository repository.PurchaseRepository, productRepository repository.ProductRepository, walletClient wallet.Client, currency string) *PurchaseUseCases {
	return &PurchaseUseCases{
		purchaseRepository: purchaseRepository,
		productRepository:  productRepository,
		walletClient:       walletClient,
		currency:           currency,
	}
}

func (u *PurchaseUseCases) Create(ctx context.Context, userID int, role string, data *domain.CreatePurchase) (*domain.Purchase, error) {
	if data.Quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be a positive integer", domain.ErrInvalidData)
	}

	product, err := u.productRepository.GetById(ctx, data.ProductID)
	if err != nil {
		return nil, err
	}

	subtotal := product.Price * domain.Cents(data.Quantity)
	var discount domain.Cents
	if domain.IsDiscountedRole(role) {
		discount = subtotal * 10 / 100
	}
	total := subtotal - discount

	reference := fmt.Sprintf("purchase-%d-%d", userID, time.Now().UnixNano())
	result, err := u.walletClient.Confirm(ctx, wallet.ConfirmRequest{
		UserID:    userID,
		ProductID: data.ProductID,
		Amount:    int64(total),
		Currency:  u.currency,
		Reference: reference,
	})
	if err != nil {
		return nil, err
	}
	if !result.Approved {
		return nil, fmt.Errorf("%w: %s", domain.ErrPaymentDeclined, result.Message)
	}

	walletRef := result.Reference
	if walletRef == "" {
		walletRef = reference
	}

	purchase, err := u.purchaseRepository.Create(ctx, &domain.Purchase{
		UserID:    userID,
		ProductID: data.ProductID,
		Quantity:  data.Quantity,
		UnitPrice: product.Price,
		Discount:  discount,
		Total:     total,
		WalletRef: &walletRef,
	})
	if err != nil {
		return nil, err
	}
	return purchase, nil
}

func (u *PurchaseUseCases) GetAllByUser(ctx context.Context, userID, page, pageSize int) ([]*domain.Purchase, int, error) {
	return u.purchaseRepository.GetAllByUser(ctx, userID, page, pageSize)
}
