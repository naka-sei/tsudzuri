package firebase

import (
	"context"

	firebasev4 "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"

	"github.com/naka-sei/tsudzuri/config"
)

type Authenticator interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

type client struct {
	client *auth.Client
}

// NewClient creates a new Firebase client.
func NewClient(conf *config.Config) (Authenticator, error) {
	if conf == nil || !conf.OnGoogleCloud() {
		return newLocalClient(), nil
	}

	ctx := context.Background()
	app, err := firebasev4.NewApp(ctx, nil)
	if err != nil {
		return nil, err
	}

	c, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return &client{
		client: c,
	}, nil
}

// VerifyIDToken verifies the given ID token and returns the UID.
func (c *client) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := c.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}

type localClient struct {
	defaultUID string
}

// newLocalClient creates a new local Firebase client.
func newLocalClient() Authenticator {
	return &localClient{defaultUID: "local-dev-user"}
}

// VerifyIDToken verifies the given ID token and returns the UID.
func (c *localClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	// In local environment, treat the provided idToken as the UID so callers can simulate
	// specific users. Fallback to a deterministic mock UID if the header is empty.
	uid := idToken
	if uid == "" {
		uid = c.defaultUID
	}

	return &auth.Token{
		UID: uid,
	}, nil
}
