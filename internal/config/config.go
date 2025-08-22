// Package config provides configuration management for the SpamAssassin MCP server.
//
// This package handles loading configuration from multiple sources with a clear
// precedence order and security-first defaults. Configuration can be loaded from:
//   1. YAML configuration files
//   2. Environment variables (with SA_MCP_ prefix)
//   3. Built-in secure defaults
//
// The configuration system includes validation and type safety to prevent
// misconfigurations that could compromise security or stability.
//
// Security considerations:
//   - All timeouts have reasonable defaults to prevent resource exhaustion
//   - Rate limits are configured to prevent abuse while allowing legitimate use
//   - Email size limits prevent memory exhaustion attacks
//   - Validation can be enabled/disabled for different environments
package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	SpamAssassin SpamAssassinConfig `mapstructure:"spamassassin"`
	Security     SecurityConfig     `mapstructure:"security"`
	LogLevel     string             `mapstructure:"log_level"`
}

type ServerConfig struct {
	BindAddr string        `mapstructure:"bind_addr"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type SpamAssassinConfig struct {
	Host      string        `mapstructure:"host"`
	Port      int           `mapstructure:"port"`
	Timeout   time.Duration `mapstructure:"timeout"`
	Threshold float64       `mapstructure:"threshold"`
}

type SecurityConfig struct {
	MaxEmailSize      int64           `mapstructure:"max_email_size"`
	RateLimiting      RateLimit       `mapstructure:"rate_limiting"`
	AllowedSenders    []string        `mapstructure:"allowed_senders"`
	BlockedDomains    []string        `mapstructure:"blocked_domains"`
	ScanTimeout       time.Duration   `mapstructure:"scan_timeout"`
	ValidationEnabled bool            `mapstructure:"validation_enabled"`
}

type RateLimit struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	BurstSize        int `mapstructure:"burst_size"`
}

func Load() (*Config, error) {
	viper.SetDefault("server.bind_addr", "0.0.0.0:8080")
	viper.SetDefault("server.timeout", "30s")
	viper.SetDefault("spamassassin.host", "localhost")
	viper.SetDefault("spamassassin.port", 783)
	viper.SetDefault("spamassassin.timeout", "30s")
	viper.SetDefault("spamassassin.threshold", 5.0)
	viper.SetDefault("security.max_email_size", 10*1024*1024) // 10MB
	viper.SetDefault("security.rate_limiting.requests_per_minute", 60)
	viper.SetDefault("security.rate_limiting.burst_size", 10)
	viper.SetDefault("security.scan_timeout", "60s")
	viper.SetDefault("security.validation_enabled", true)
	viper.SetDefault("log_level", "info")

	// Environment variables
	viper.SetEnvPrefix("SA_MCP")
	viper.AutomaticEnv()

	// Read config file if it exists
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/spamassassin-mcp")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// Config file not found, use defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}