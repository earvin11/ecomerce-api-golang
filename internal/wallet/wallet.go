package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ecomerce-api/internal/domain"
)

type ConfirmRequest struct {
	UserID    int    `json:"user_id"`
	ProductID int    `json:"product_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Reference string `json:"reference"`
}

type ConfirmResponse struct {
	Approved  bool   `json:"approved"`
	Reference string `json:"reference"`
	Message   string `json:"message"`
}

type ConfirmResult struct {
	Approved  bool
	Reference string
	Message   string
}

type Client interface {
	Confirm(ctx context.Context, req ConfirmRequest) (*ConfirmResult, error)
}

type Config struct {
	BaseURL string
	Timeout time.Duration
}

type httpClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPClient(cfg Config) Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &httpClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *httpClient) Confirm(ctx context.Context, req ConfirmRequest) (*ConfirmResult, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("%w: wallet base url is not configured", domain.ErrWalletUnavailable)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("wallet client marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("wallet client build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrWalletUnavailable, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("wallet client read response: %w", err)
	}

	var parsed ConfirmResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("%w: malformed response from wallet (status %d)", domain.ErrWalletUnavailable, resp.StatusCode)
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("%w: wallet server error (status %d)", domain.ErrWalletUnavailable, resp.StatusCode)
	}

	return &ConfirmResult{
		Approved:  resp.StatusCode >= 200 && resp.StatusCode < 300 && parsed.Approved,
		Reference: parsed.Reference,
		Message:   parsed.Message,
	}, nil
}
