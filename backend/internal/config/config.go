// Package config reads the environment. No secret has a default; nothing here logs a value.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration.
type Config struct {
	Env      string // "development" | "production"
	Server   ServerConfig
	Database *DatabaseConfig
	Auth0    Auth0Config
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         string
	FrontendURL  string
	ReadTimeout  int
	WriteTimeout int
}

// Auth0Config holds the Auth0 settings used to verify access tokens.
type Auth0Config struct {
	Domain   string
	Audience string
}

// Validate reports which required Auth0 settings are empty: the API cannot verify tokens without them.
func (c Auth0Config) Validate() error {
	var missing []string
	if strings.TrimSpace(c.Domain) == "" {
		missing = append(missing, "AUTH0_DOMAIN")
	}
	if strings.TrimSpace(c.Audience) == "" {
		missing = append(missing, "AUTH0_AUDIENCE")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required Auth0 settings: %s", strings.Join(missing, ", "))
	}
	return nil
}

// Load creates the configuration from environment variables.
func Load() *Config {
	return &Config{
		Env: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:4200"),
			ReadTimeout:  15,
			WriteTimeout: 30,
		},
		Database: LoadDatabaseConfig(),
		Auth0: Auth0Config{
			Domain:   strings.TrimSpace(getEnv("AUTH0_DOMAIN", "")),
			Audience: strings.TrimSpace(getEnv("AUTH0_AUDIENCE", "")),
		},
	}
}

// Validate refuses unsafe production settings (plain-text database connection, http frontend).
func (c *Config) Validate() error {
	if err := c.Auth0.Validate(); err != nil {
		return err
	}
	if c.Env == "production" {
		if !strings.HasPrefix(c.Server.FrontendURL, "https://") {
			return fmt.Errorf("FRONTEND_URL must be https in production")
		}
		if c.Database.SSLMode == "disable" {
			return fmt.Errorf("DB_SSLMODE must not be 'disable' in production (use require, verify-ca or verify-full)")
		}
		if c.Database.Password == "" {
			return fmt.Errorf("DB_PASSWORD is required in production")
		}
	}
	return nil
}
