package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/kelseyhightower/envconfig"
)

var conf *Config

// Config represents the application configuration.
type Config struct {
	// Port is the port number the server listens on.
	Port uint `envconfig:"PORT" default:"8080"`

	// GRPCPort is the port number the gRPC server listens on.
	GRPCPort uint `envconfig:"GRPC_PORT" default:"9090"`

	// GoogleCloudProject is the Google Cloud project name.
	GoogleCloudProject string `envconfig:"GOOGLE_CLOUD_PROJECT" default:"tsudzuri-dev"`

	// IsDebugMode indicates whether the application is running in debug mode.
	IsDebugMode bool `envconfig:"IS_DEBUG_MODE" default:"false"`

	// EnableStdoutTraceExporter enables the OpenTelemetry stdout exporter for local debugging.
	EnableStdoutTraceExporter bool `envconfig:"ENABLE_STDOUT_TRACE_EXPORTER" default:"false"`

	// TsuzduriDatabaseDSNSMKey is the Secret Manager key for the Tsudzuri database DSN.
	TsuzduriDatabaseDSNSMKey string `envconfig:"TSUDZURI_DATABASE_DSN_SM_KEY" default:"tsudzuri_database-dsn"`

	// TsudzuriDatabaseDSN is the Tsudzuri database DSN.
	TsudzuriDatabaseDSN string `envconfig:"TSUDZURI_DATABASE_DSN"`
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
	if conf.OnGoogleCloud() {
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
	// Determine project to read secrets from. Prefer explicit config value,
	// but allow environment variable override.
	project := c.GoogleCloudProject
	if p := os.Getenv("GOOGLE_CLOUD_PROJECT"); p != "" {
		project = p
	}

	if project == "" {
		return fmt.Errorf("google cloud project is not set")
	}

	// If no secret key is configured, nothing to do.
	smKey := strings.TrimSpace(c.TsuzduriDatabaseDSNSMKey)
	if smKey == "" {
		return nil
	}

	// Use shared helper to fetch secret value
	val, err := FetchSMSecret(ctx, project, smKey)
	if err != nil {
		return err
	}
	c.TsudzuriDatabaseDSN = strings.TrimSpace(val)
	return nil
}

// FetchSMSecret reads the secret payload for the given project and secret key from
// Secret Manager and returns it as a string. Caller should trim or parse the value
// as needed.
func FetchSMSecret(ctx context.Context, project, key string) (string, error) {
	if project == "" {
		return "", fmt.Errorf("google cloud project is not set")
	}
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("secret key is empty")
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("creating secretmanager client: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: closing secretmanager client: %v\n", err)
		}
	}()

	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", project, key)
	req := &secretmanagerpb.AccessSecretVersionRequest{Name: name}
	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("accessing secret %s: %w", name, err)
	}

	if resp.Payload == nil || len(resp.Payload.Data) == 0 {
		return "", fmt.Errorf("secret %s has empty payload", name)
	}

	return string(resp.Payload.Data), nil
}

// OnGoogleCloud determines if the application is running on Google Cloud.
// It checks for the presence of the "K_SERVICE" environment variable, which is set in Cloud Run.
func (c *Config) OnGoogleCloud() bool {
	hasCloudRunVar := os.Getenv("K_SERVICE") != ""
	return hasCloudRunVar
}
