package gh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/google/go-github/v51/github"
	"github.com/gregjones/httpcache"
	"github.com/rs/zerolog/log"
)

// caches an instance of the client for each authenticated scope
var clients map[string]Client

type Client struct {
	client *github.Client
	config.RunnerScope
}

func NewClient(ctx context.Context, scope config.RunnerScope) (*Client, error) {
	if clients == nil {
		clients = map[string]Client{}
	}

	if client, ok := clients[scope.String()]; ok {
		return &client, nil
	}

	oauthTransport, err := githubAuthTransport(ctx)
	if err != nil {
		return nil, err
	}

	cache := httpcache.NewTransport(httpcache.NewMemoryCache())
	cache.Transport = oauthTransport

	logging := loggingTransport{Transport: cache}

	httpClient := http.Client{Transport: logging}
	githubClient := github.NewClient(&httpClient)

	client := Client{client: githubClient, RunnerScope: scope}
	clients[scope.String()] = client
	return &client, nil
}

func (c Client) CreateRegistrationToken(ctx context.Context) (token *github.RegistrationToken, err error) {
	log.Debug().
		Bool("is_org", c.RunnerScope.IsOrg()).
		Bool("is_repo", c.RunnerScope.IsRepo()).
		Str("org", c.RunnerScope.Organisation).
		Str("repo", c.RunnerScope.Repository).
		Send()

	if c.RunnerScope.IsOrg() {
		token, _, err = c.client.Actions.CreateOrganizationRegistrationToken(ctx, c.RunnerScope.Organisation)
	} else if c.RunnerScope.IsRepo() {
		owner, repo, found := c.RunnerScope.GetRepo()
		if !found {
			return nil, fmt.Errorf("repository must be formatted as <owner>/<name>, got: %s", c.RunnerScope.Repository)
		}

		token, _, err = c.client.Actions.CreateRegistrationToken(ctx, owner, repo)
	} else {
		return nil, fmt.Errorf("no github authentication scope available")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create registration token: %s", err)
	}

	return token, err
}
