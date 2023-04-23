package gh

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/axatol/actions-runner-broker/pkg/config"
	"github.com/google/go-github/v51/github"
	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"
)

// caches an instance of the client for each authenticated scope
var clients map[string]Client

type Client struct {
	client *github.Client
	config.RunnerScope
}

func NewClient(ctx context.Context, scope string) *Client {
	if clients == nil {
		clients = map[string]Client{}
	}

	if client, ok := clients[scope]; ok {
		return &client
	}

	token := oauth2.Token{AccessToken: config.GithubToken.String()}
	oauth2Client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&token))

	cache := httpcache.NewTransport(httpcache.NewMemoryCache())
	cache.Transport = oauth2Client.Transport

	logging := LoggingTransport{Transport: cache}

	httpClient := http.Client{Transport: logging}
	githubClient := github.NewClient(&httpClient)

	client := Client{client: githubClient}
	clients[scope] = client
	return &client
}

func (c Client) CreateRegistrationToken(ctx context.Context) (token *github.RegistrationToken, err error) {
	if c.RunnerScope.Organisation != "" {
		token, _, err = c.client.Actions.CreateOrganizationRegistrationToken(ctx, c.RunnerScope.Organisation)
	} else if c.RunnerScope.Repository != "" {
		owner, repo, found := strings.Cut(c.RunnerScope.Repository, "/")
		if !found {
			return nil, fmt.Errorf("repository must be formatted as <owner>/<name>, got: %s", c.RunnerScope.Repository)
		}

		token, _, err = c.client.Actions.CreateRegistrationToken(ctx, owner, repo)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create registration token: %s", err)
	}

	return token, err
}
