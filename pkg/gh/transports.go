package gh

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gregjones/httpcache"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const (
	headerRateLimitLimit     = "x-ratelimit-limit"
	headerRateLimitRemaining = "x-ratelimit-remaining"
	headerRateLimitUsed      = "x-ratelimit-used"
	headerRateLimitReset     = "x-ratelimit-reset"
)

type loggingTransport struct{ Transport http.RoundTripper }

func (t loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.Transport.RoundTrip(req)
	t.log(req, res, err)
	return res, err
}

func (t loggingTransport) log(req *http.Request, res *http.Response, err error) {
	event := log.Info()
	if err != nil {
		event = log.Error().Err(err)
	}

	if res != nil {
		event = event.
			Int("status_code", res.StatusCode).
			Bool("cache_hit", res.Header.Get(httpcache.XFromCache) == "1").
			Str("rate_limit_remaining", res.Header.Get(headerRateLimitRemaining))
	}

	event.
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Send()
}

func githubAuthTransport(ctx context.Context) (transport http.RoundTripper, err error) {
	if config.GithubToken.Value() != "" {
		token := oauth2.Token{AccessToken: config.GithubToken.String()}
		oauth2Client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&token))
		return oauth2Client.Transport, nil
	}

	if _, existsErr := os.Stat(config.GithubAppPrivateKey.String()); existsErr == nil {
		transport, err = ghinstallation.NewKeyFromFile(
			http.DefaultTransport,
			config.GithubAppID.Value(),
			config.GithubAppInstallationID.Value(),
			config.GithubAppPrivateKey.Value(),
		)
	} else {
		transport, err = ghinstallation.New(
			http.DefaultTransport,
			config.GithubAppID.Value(),
			config.GithubAppInstallationID.Value(),
			[]byte(config.GithubAppPrivateKey.Value()),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("github app authentication failed: %s", err)
	}

	return transport, nil
}
