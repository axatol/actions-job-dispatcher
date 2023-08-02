package gh

import (
	"context"
	"net/http"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"github.com/google/go-github/v51/github"
	"github.com/gregjones/httpcache"
)

// caches an instance of the client for each authenticated scope
var clients = map[string]Client{}

type Client struct {
	client *github.Client
	scope  config.Scope
}

func GetClient(ctx context.Context, scope config.Scope) (*Client, error) {
	if client, ok := clients[scope.String()]; ok {
		return &client, nil
	}

	authTransport, err := config.Github.Transport(ctx)
	if err != nil {
		return nil, err
	}

	cache := httpcache.NewTransport(httpcache.NewMemoryCache())
	cache.Transport = authTransport

	logging := loggingTransport{Transport: cache}

	httpClient := http.Client{Transport: logging}
	githubClient := github.NewClient(&httpClient)

	client := Client{client: githubClient, scope: scope}
	clients[scope.String()] = client
	return &client, nil
}
