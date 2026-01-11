package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	OIDC         OIDCConfig
	OpenFGA      OpenFGAConfig
	MultiTenant  MultiTenantConfig
	OpenLineage  OpenLineageConfig
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// OIDCConfig contains OIDC provider configuration
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// OpenFGAConfig contains OpenFGA authorization configuration
type OpenFGAConfig struct {
	StoreID   string
	APIHost   string
	AuthModel string
}

// MultiTenantConfig contains multi-tenancy settings
type MultiTenantConfig struct {
	Enabled        bool
	IsolationLevel string // namespace, database, schema
}

// OpenLineageConfig contains OpenLineage integration settings
type OpenLineageConfig struct {
	Enabled  bool
	Endpoint string
	Namespace string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Database: getEnv("DB_NAME", "opendq"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		OIDC: OIDCConfig{
			Issuer:       getEnv("OIDC_ISSUER", ""),
			ClientID:     getEnv("OIDC_CLIENT_ID", ""),
			ClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("OIDC_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		},
		OpenFGA: OpenFGAConfig{
			StoreID:   getEnv("OPENFGA_STORE_ID", ""),
			APIHost:   getEnv("OPENFGA_API_HOST", "http://localhost:8081"),
			AuthModel: getEnv("OPENFGA_AUTH_MODEL", ""),
		},
		MultiTenant: MultiTenantConfig{
			Enabled:        getEnvAsBool("MULTITENANT_ENABLED", true),
			IsolationLevel: getEnv("MULTITENANT_ISOLATION", "namespace"),
		},
		OpenLineage: OpenLineageConfig{
			Enabled:   getEnvAsBool("OPENLINEAGE_ENABLED", true),
			Endpoint:  getEnv("OPENLINEAGE_ENDPOINT", "http://localhost:5000"),
			Namespace: getEnv("OPENLINEAGE_NAMESPACE", "opendq"),
		},
	}

	return cfg, nil
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
