package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"ecomerce-api/internal/domain"
)

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenService(secret string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *TokenService) AccessTTL() time.Duration {
	return s.accessTTL
}

func (s *TokenService) SignAccess(claims domain.TokenClaims) (string, error) {
	return s.sign(claims, s.accessTTL)
}

func (s *TokenService) SignRefresh(claims domain.TokenClaims) (string, error) {
	return s.sign(claims, s.refreshTTL)
}

func (s *TokenService) ParseAccess(token string) (*domain.TokenClaims, error) {
	return s.parse(token)
}

func (s *TokenService) ParseRefresh(token string) (*domain.TokenClaims, error) {
	return s.parse(token)
}

func (s *TokenService) sign(claims domain.TokenClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})
	return token.SignedString(s.secret)
}

func (s *TokenService) parse(token string) (*domain.TokenClaims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	return &domain.TokenClaims{UserID: claims.UserID, Username: claims.Username, Role: claims.Role}, nil
}
