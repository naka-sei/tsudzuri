package config

import (
	"context"
	"os"

	"github.com/kelseyhightower/envconfig"
)

var conf *Config

// Config represents the application configuration.
type Config struct {
	// Port is the port number the server listens on.
	Port uint `envconfig:"PORT" default:"8080"`

	// GoogleCloudProject is the Google Cloud project name.
	GoogleCloudProject string `envconfig:"GOOGLE_CLOUD_PROJECT" default:"tsudzuri-dev"`

	// IsDebugMode indicates whether the application is running in debug mode.
	IsDebugMode bool `envconfig:"IS_DEBUG_MODE" default:"false"`

	// TsuzduriDatabaseDSNSMKey is the Secret Manager key for the Tsudzuri database DSN.
	TsuzduriDatabaseDSNSMKey string `envconfig:"TSUDZURI_DATABASE_DSN_SM_KEY" default:"tsudzuri_database-dsn"`

	// TsudzuriDatabaseDSN is the Tsudzuri database DSN.
	TsudzuriDatabaseDSN string `envconfig:"TSUDZURI_DATABASE_DSN"`

	// TestDatabaseDSN is the test database DSN.
	TestDatabaseDSN string `envconfig:"TEST_DATABASE_DSN"`
}

// Load loads the configuration.
// On Google Cloud most environments, it loads from Secret Manager first, then overrides with environment variables.
// On other environments, it loads only from environment variables.
func Load(ctx context.Context) (*Config, error) {
	conf = new(Config)

	// Load from Secret Manager
	if err := conf.loadFromEnv(); err != nil {
		return nil, err
	}

	// Override with environment variables
	if onGoogleCloud() {
		if err := conf.loadFromSM(ctx); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

// loadFromEnv loads the configuration from environment variables.
func (c *Config) loadFromEnv() error {
	if err := envconfig.Process("", c); err != nil {
		return err
	}
	return nil
}

// loadFromSM loads the configuration from Secret Manager.
func (c *Config) loadFromSM(ctx context.Context) error {
	// Implement loading from Secret Manager here.
	return nil
}

// onGoogleCloud determines if the application is running on Google Cloud.
// It checks for the presence of the "K_SERVICE" environment variable, which is set in Cloud Run.
func onGoogleCloud() bool {
	hasCloudRunVar := os.Getenv("K_SERVICE") != ""
	return hasCloudRunVar
}
