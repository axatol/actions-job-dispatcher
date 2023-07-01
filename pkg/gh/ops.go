package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v51/github"
)

func (c *Client) ListRunners(ctx context.Context) ([]*github.Runner, error) {
	var (
		opts       *github.ListOptions
		allRunners []*github.Runner
		runners    *github.Runners
		resp       *github.Response
		err        error
	)

	for {
		if c.scope.IsOrg {
			runners, resp, err = c.client.Actions.ListOrganizationRunners(ctx, c.scope.String(), opts)
		} else {
			runners, resp, err = c.client.Actions.ListRunners(ctx, c.scope.Owner, c.scope.Repository, opts)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list runners: %s", err)
		}

		allRunners = append(allRunners, runners.Runners...)
		if resp.NextPage < 1 {
			break
		}

		opts.Page = resp.NextPage
	}

	return allRunners, nil
}

func (c *Client) DescribeScope(ctx context.Context) (string, error) {
	if c.scope.IsOrg {
		org, _, err := c.client.Organizations.Get(ctx, c.scope.String())
		if err != nil {
			return "", fmt.Errorf("failed to get organisation %s: %s", c.scope.String(), err)
		}

		return org.GetHTMLURL(), nil
	} else {
		repo, _, err := c.client.Repositories.Get(ctx, c.scope.Owner, c.scope.Repository)
		if err != nil {
			return "", fmt.Errorf("failed to get repository %s: %s", c.scope.String(), err)
		}

		return repo.GetHTMLURL(), nil
	}
}

func (c Client) CreateRegistrationToken(ctx context.Context) (*github.RegistrationToken, error) {
	if c.scope.IsOrg {
		token, _, err := c.client.Actions.CreateOrganizationRegistrationToken(ctx, c.scope.String())
		if err != nil {
			return nil, fmt.Errorf("failed to create registration token for organisation %s: %s", c.scope.String(), err)
		}

		return token, nil
	} else {
		token, _, err := c.client.Actions.CreateRegistrationToken(ctx, c.scope.Owner, c.scope.Repository)
		if err != nil {
			return nil, fmt.Errorf("failed to create registration token for repository %s: %s", c.scope.String(), err)
		}

		return token, nil
	}
}
