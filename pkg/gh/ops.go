package gh

import (
	"context"
	"fmt"

	"github.com/axatol/actions-job-dispatcher/pkg/cache"
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
			return nil, fmt.Errorf("failed to list runners for %s: %s", c.scope.String(), err)
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
	var (
		scope interface{ GetHTMLURL() string }
		err   error
	)

	if c.scope.IsOrg {
		scope, _, err = c.client.Organizations.Get(ctx, c.scope.String())
	} else {
		scope, _, err = c.client.Repositories.Get(ctx, c.scope.Owner, c.scope.Repository)
	}

	if err != nil {
		return "", fmt.Errorf("failed to describe scope %s: %s", c.scope.String(), err)
	}

	return scope.GetHTMLURL(), nil
}

func (c Client) CreateRegistrationToken(ctx context.Context) (*github.RegistrationToken, error) {
	var (
		token *github.RegistrationToken
		err   error
	)

	if c.scope.IsOrg {
		token, _, err = c.client.Actions.CreateOrganizationRegistrationToken(ctx, c.scope.String())
	} else {
		token, _, err = c.client.Actions.CreateRegistrationToken(ctx, c.scope.Owner, c.scope.Repository)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create registration token for %s: %s", c.scope.String(), err)
	}

	return token, nil
}

func (c Client) DescribeWorkflowJob(ctx context.Context, meta *cache.WorkflowJobMeta) (*github.WorkflowJob, error) {
	job, _, err := c.client.Actions.GetWorkflowJobByID(ctx, meta.Owner, meta.Repository, meta.WorkflowJobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow job by id for %s: %s", c.scope.String(), err)
	}

	return job, nil
}
