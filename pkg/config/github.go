package config

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"golang.org/x/oauth2"
)

type GithubConfig struct {
	Token             string `yaml:"token"`
	AppID             int64  `yaml:"app_id"`
	AppInstallationID int64  `yaml:"app_installation_id"`
	AppPrivateKey     string `yaml:"app_private_key"`
	AppPrivateKeyFile string `yaml:"app_private_key_file"`
}

func (c GithubConfig) IsToken() bool {
	return c.Token != ""
}

func (c GithubConfig) IsApp() bool {
	return c.AppID > 0 && c.AppInstallationID > 0 && (c.AppPrivateKey != "" || c.AppPrivateKeyFile != "")
}

func (c GithubConfig) Validate() error {
	if !c.IsToken() && !c.IsApp() {
		return fmt.Errorf("must specify token or app details")
	}

	return nil
}

func (c GithubConfig) Transport(ctx context.Context) (http.RoundTripper, error) {
	if c.IsToken() {
		token := oauth2.Token{AccessToken: c.Token}
		oauth2Client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&token))
		return oauth2Client.Transport, nil
	}

	if c.AppPrivateKeyFile != "" {
		transport, err := ghinstallation.NewKeyFromFile(
			http.DefaultTransport,
			c.AppID,
			c.AppInstallationID,
			c.AppPrivateKeyFile,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with github private key from file: %s", err)
		}

		return transport, nil
	}

	transport, err := ghinstallation.New(
		http.DefaultTransport,
		c.AppID,
		c.AppInstallationID,
		[]byte(c.AppPrivateKey),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with github raw private key: %s", err)
	}

	return transport, nil
}
