package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort        string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxConns        int32
	JWTSecret         string
	JWTAccessTTL      time.Duration
	JWTRefreshTTL     time.Duration
	WalletAPIURL      string
	WalletAPITimeout  time.Duration
	WalletCurrency    string
	SeedAdminUsername string
	SeedAdminEmail    string
	SeedAdminPassword string
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		ReadTimeout:       getDurationEnv("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:      getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:       getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "ecommerce"),
		DBSSLMode:         getEnv("DB_SSL_MODE", "disable"),
		DBMaxConns:        int32(getIntEnv("DB_MAX_CONNS", 10)),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTAccessTTL:      getDurationEnv("JWT_ACCESS_TTL", 30*time.Minute),
		JWTRefreshTTL:     getDurationEnv("JWT_REFRESH_TTL", 720*time.Hour),
		WalletAPIURL:      getEnv("WALLET_API_URL", ""),
		WalletAPITimeout:  getDurationEnv("WALLET_API_TIMEOUT", 5*time.Second),
		WalletCurrency:    getEnv("WALLET_CURRENCY", "USD"),
		SeedAdminUsername: getEnv("SEED_ADMIN_USERNAME", ""),
		SeedAdminEmail:    getEnv("SEED_ADMIN_EMAIL", ""),
		SeedAdminPassword: getEnv("SEED_ADMIN_PASSWORD", ""),
	}

	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("config: DB_HOST, DB_USER and DB_NAME are required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET is required (set a long random value)")
	}
	if cfg.WalletAPIURL == "" {
		return nil, fmt.Errorf("config: WALLET_API_URL is required (external wallet confirmation service)")
	}
	set := 0
	for _, v := range []string{cfg.SeedAdminUsername, cfg.SeedAdminEmail, cfg.SeedAdminPassword} {
		if v != "" {
			set++
		}
	}
	if set > 0 && set < 3 {
		return nil, fmt.Errorf("config: SEED_ADMIN_USERNAME, SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD must be set together")
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
