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

type RateLimit struct {
	Limit     string `json:"-"`
	Remaining string `json:"remaining"`
	Used      string `json:"-"`
	Reset     string `json:"-"`
}

var usage = map[string]RateLimit{}

type loggingTransport struct {
	Transport http.RoundTripper
	scope     config.Scope
}

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
		usage[t.scope.String()] = RateLimit{
			Limit:     res.Header.Get(headerRateLimitLimit),
			Remaining: res.Header.Get(headerRateLimitRemaining),
			Used:      res.Header.Get(headerRateLimitUsed),
			Reset:     res.Header.Get(headerRateLimitReset),
		}

		event = event.
			Int("status_code", res.StatusCode).
			Bool("cache_hit", res.Header.Get(httpcache.XFromCache) == "1").
			Str("rate_limit_remaining", usage[t.scope.String()].Remaining)
	}

	event.
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Send()
}

func githubAuthTransport(ctx context.Context) (transport http.RoundTripper, err error) {
	if config.Github.Token != "" {
		token := oauth2.Token{AccessToken: config.Github.Token}
		oauth2Client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&token))
		return oauth2Client.Transport, nil
	}

	if _, existsErr := os.Stat(config.Github.AppPrivateKey); existsErr == nil {
		transport, err = ghinstallation.NewKeyFromFile(
			http.DefaultTransport,
			config.Github.AppID,
			config.Github.AppInstallationID,
			config.Github.AppPrivateKey,
		)
	} else {
		transport, err = ghinstallation.New(
			http.DefaultTransport,
			config.Github.AppID,
			config.Github.AppInstallationID,
			[]byte(config.Github.AppPrivateKey),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("github app authentication failed: %s", err)
	}

	return transport, nil
}
