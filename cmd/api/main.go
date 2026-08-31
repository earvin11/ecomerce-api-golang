package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecomerce-api/internal/auth"
	"ecomerce-api/internal/config"
	"ecomerce-api/internal/handlers"
	"ecomerce-api/internal/repository"
	usecases "ecomerce-api/internal/use_cases"
	"ecomerce-api/internal/wallet"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := repository.NewPool(ctx, cfg.DSN(), cfg.DBMaxConns)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := repository.EnsureSchema(ctx, pool); err != nil {
		slog.Error("failed to ensure schema", "error", err)
		os.Exit(1)
	}

	if err := repository.SeedRoles(ctx, pool); err != nil {
		slog.Error("failed to seed roles", "error", err)
		os.Exit(1)
	}

	if cfg.SeedAdminUsername != "" && cfg.SeedAdminEmail != "" && cfg.SeedAdminPassword != "" {
		if err := repository.SeedAdminUser(ctx, pool, cfg.SeedAdminEmail, cfg.SeedAdminPassword, cfg.SeedAdminUsername); err != nil {
			slog.Error("failed to seed admin user", "error", err)
			os.Exit(1)
		}
	}

	categoryRepo := repository.NewCategoryRepository(pool)
	productRepo := repository.NewProductRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	purchaseRepo := repository.NewPurchaseRepository(pool)

	categoryUseCases := usecases.NewCategoryUseCases(categoryRepo)
	productUseCases := usecases.NewProductUseCases(productRepo)
	roleUseCases := usecases.NewRoleUseCases(roleRepo)
	userUseCases := usecases.NewUserUseCases(userRepo)

	walletClient := wallet.NewHTTPClient(wallet.Config{BaseURL: cfg.WalletAPIURL, Timeout: cfg.WalletAPITimeout})
	purchaseUseCases := usecases.NewPurchaseUseCases(purchaseRepo, productRepo, walletClient, cfg.WalletCurrency)

	tokenService := auth.NewTokenService(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	authUseCases := usecases.NewAuthUseCases(userRepo, tokenService)

	categoryHandler := handlers.NewCategoryHandler(categoryUseCases)
	productHandler := handlers.NewProductHandler(productUseCases)
	roleHandler := handlers.NewRoleHandler(roleUseCases)
	userHandler := handlers.NewUserHandler(userUseCases)
	authHandler := handlers.NewAuthHandler(authUseCases)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseUseCases)
	authMiddleware := handlers.NewAuthMiddleware(tokenService)

	handler := handlers.NewHandler(categoryHandler, productHandler, roleHandler, userHandler, authHandler, purchaseHandler, authMiddleware, pool)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      handler.Router(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		slog.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
