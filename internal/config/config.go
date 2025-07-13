// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Enrichment  EnrichmentConfig  `yaml:"enrichment"`
	Rules       RulesConfig       `yaml:"rules"`
	Enforcement EnforcementConfig `yaml:"enforcement"`
	Storage     StorageConfig     `yaml:"storage"`
	Logging     LoggingConfig     `yaml:"logging"`
	DryRun      bool              `yaml:"dry_run"`
}

type MonitoringConfig struct {
	Sources SourcesConfig `yaml:"sources"`
}

type SourcesConfig struct {
	Certstream CertstreamConfig `yaml:"certstream"`
}

type CertstreamConfig struct {
	Enabled  bool     `yaml:"enabled"`
	URL      string   `yaml:"url"`
	Keywords []string `yaml:"keywords"`
}

type EnrichmentConfig struct {
	HTMLContent HTMLContentConfig `yaml:"html_content"`
	Favicon     FaviconConfig     `yaml:"favicon"`
}

type HTMLContentConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Timeout   string `yaml:"timeout"`
	UserAgent string `yaml:"user_agent"`
}

type FaviconConfig struct {
	Enabled bool   `yaml:"enabled"`
	Timeout string `yaml:"timeout"`
}

type RulesConfig struct {
	FaviconSimilarity FaviconSimilarityConfig `yaml:"favicon_similarity"`
}

type FaviconSimilarityConfig struct {
	Enabled           bool              `yaml:"enabled"`
	Threshold         float64           `yaml:"threshold"`
	ReferenceFavicons map[string]string `yaml:"reference_favicons"`
}

type EnforcementConfig struct {
	EmailAbuse EmailAbuseConfig `yaml:"email_abuse"`
	Logger     LoggerConfig     `yaml:"logger"`
}

type EmailAbuseConfig struct {
	Enabled bool       `yaml:"enabled"`
	SMTP    SMTPConfig `yaml:"smtp"`
	From    string     `yaml:"from"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type LoggerConfig struct {
	Enabled bool `yaml:"enabled"`
}

type StorageConfig struct {
	Type string `yaml:"type"` // memory, sqlite, postgres
}

type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // text, json
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(filename string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", filename)
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the config
	expandedData := os.ExpandEnv(string(data))

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	config.applyDefaults()

	// Validate config
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// applyDefaults sets default values for missing configuration
func (c *Config) applyDefaults() {
	// Monitoring defaults
	if c.Monitoring.Sources.Certstream.URL == "" {
		c.Monitoring.Sources.Certstream.URL = "wss://certstream.calidog.io/"
	}

	// Enrichment defaults
	if c.Enrichment.HTMLContent.Timeout == "" {
		c.Enrichment.HTMLContent.Timeout = "10s"
	}
	if c.Enrichment.HTMLContent.UserAgent == "" {
		c.Enrichment.HTMLContent.UserAgent = "OpenBPL/1.0"
	}
	if c.Enrichment.Favicon.Timeout == "" {
		c.Enrichment.Favicon.Timeout = "5s"
	}

	// Rules defaults
	if c.Rules.FaviconSimilarity.Threshold == 0 {
		c.Rules.FaviconSimilarity.Threshold = 0.85
	}

	// Storage defaults
	if c.Storage.Type == "" {
		c.Storage.Type = "memory"
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}

	// SMTP defaults
	if c.Enforcement.EmailAbuse.SMTP.Port == 0 {
		c.Enforcement.EmailAbuse.SMTP.Port = 587
	}
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	// Check storage type
	validStorageTypes := map[string]bool{
		"memory":   true,
		"sqlite":   true,
		"postgres": true,
	}
	if !validStorageTypes[c.Storage.Type] {
		return fmt.Errorf("invalid storage type: %s (must be: memory, sqlite, postgres)", c.Storage.Type)
	}

	// Check logging level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be: debug, info, warn, error)", c.Logging.Level)
	}

	// Check favicon similarity threshold
	if c.Rules.FaviconSimilarity.Enabled {
		if c.Rules.FaviconSimilarity.Threshold < 0 || c.Rules.FaviconSimilarity.Threshold > 1 {
			return fmt.Errorf("favicon similarity threshold must be between 0 and 1, got: %f", c.Rules.FaviconSimilarity.Threshold)
		}
	}

	// Check SMTP configuration if email enforcement is enabled
	if c.Enforcement.EmailAbuse.Enabled {
		if c.Enforcement.EmailAbuse.SMTP.Host == "" {
			return fmt.Errorf("SMTP host is required when email enforcement is enabled")
		}
		if c.Enforcement.EmailAbuse.From == "" {
			return fmt.Errorf("email from address is required when email enforcement is enabled")
		}
	}

	return nil
}

// CreateSampleConfig creates a sample configuration file
func CreateSampleConfig(filename string) error {
	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("configuration file already exists: %s", filename)
	}

	sampleConfig := `# OpenBPL Configuration
# This is a sample configuration for the OpenBPL monitoring system

# Monitoring configuration
monitoring:
  sources:
    certstream:
      enabled: true
      url: "wss://certstream.calidog.io/"
      keywords:
        - "paypal"
        - "amazon" 
        - "microsoft"
        - "apple"
        - "google"

# Enrichment settings  
enrichment:
  html_content:
    enabled: true
    timeout: "10s"
    user_agent: "OpenBPL/1.0"
  favicon:
    enabled: true
    timeout: "5s"

# Detection rules
rules:
  favicon_similarity:
    enabled: true
    threshold: 0.85
    reference_favicons:
      paypal: "https://www.paypal.com/favicon.ico"
      amazon: "https://www.amazon.com/favicon.ico"
      microsoft: "https://www.microsoft.com/favicon.ico"
      apple: "https://www.apple.com/favicon.ico"
      google: "https://www.google.com/favicon.ico"

# Enforcement actions
enforcement:
  email_abuse:
    enabled: true
    smtp:
      host: "smtp.gmail.com"
      port: 587
      username: "alerts@yourdomain.com"
      password: "${SMTP_PASSWORD}"
    from: "OpenBPL <alerts@yourdomain.com>"
  logger:
    enabled: true

# Storage configuration
storage:
  type: "memory"  # Options: memory, sqlite, postgres
  
# Logging
logging:
  level: "info"
  format: "text"

# Run in dry-run mode (no enforcement actions will be taken)
dry_run: false
`

	if err := os.WriteFile(filename, []byte(sampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
