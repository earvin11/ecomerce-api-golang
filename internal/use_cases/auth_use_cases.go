package usecases

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"ecomerce-api/internal/auth"
	"ecomerce-api/internal/domain"
	"ecomerce-api/internal/repository"
)

type AuthUseCases struct {
	userRepository repository.UserRepository
	tokens         *auth.TokenService
}

func NewAuthUseCases(userRepository repository.UserRepository, tokens *auth.TokenService) *AuthUseCases {
	return &AuthUseCases{userRepository: userRepository, tokens: tokens}
}

func (u *AuthUseCases) Login(ctx context.Context, creds *domain.Credentials) (*domain.TokenPair, error) {
	user, err := u.userRepository.GetByUsername(ctx, creds.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%w: invalid username or password", domain.ErrUnauthorized)
		}
		return nil, err
	}
	if user.Role == nil {
		return nil, fmt.Errorf("%w: invalid username or password", domain.ErrUnauthorized)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		return nil, fmt.Errorf("%w: invalid username or password", domain.ErrUnauthorized)
	}
	return u.issuePair(user)
}

func (u *AuthUseCases) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	claims, err := u.tokens.ParseRefresh(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid or expired refresh token", domain.ErrUnauthorized)
	}
	user, err := u.userRepository.GetByUsername(ctx, claims.Username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%w: user no longer exists", domain.ErrUnauthorized)
		}
		return nil, err
	}
	return u.issuePair(user)
}

func (u *AuthUseCases) issuePair(user *domain.User) (*domain.TokenPair, error) {
	claims := domain.TokenClaims{Username: user.Username, Role: user.Role.Name}

	access, err := u.tokens.SignAccess(claims)
	if err != nil {
		return nil, fmt.Errorf("auth use cases sign access: %w", err)
	}
	refresh, err := u.tokens.SignRefresh(claims)
	if err != nil {
		return nil, fmt.Errorf("auth use cases sign refresh: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.tokens.AccessTTL().Seconds()),
	}, nil
}
