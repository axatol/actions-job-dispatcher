package k8s

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
)

func (c *Client) Version() (*version.Info, error) {
	version, err := c.client.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve server version: %s", err)
	}

	return version, nil
}

func Version() (*version.Info, error) {
	client, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes client: %s", err)
	}

	return client.Version()
}

func (c *Client) ListJobs(ctx context.Context) ([]batchv1.Job, error) {
	opts := metav1.ListOptions{LabelSelector: JobSelector.String()}
	jobs, err := c.client.BatchV1().Jobs(c.namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %s", err)
	}

	return jobs.Items, nil
}

func ListJobs(ctx context.Context) ([]batchv1.Job, error) {
	client, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes client: %s", err)
	}

	return client.ListJobs(ctx)
}

func (c *Client) CreateJob(ctx context.Context, job batchv1.Job) (*batchv1.Job, error) {
	response, err := c.client.BatchV1().Jobs(job.Namespace).Create(ctx, &job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job %s/%s: %s", job.Namespace, job.Name, err)
	}

	return response, nil
}

func CreateJob(ctx context.Context, job batchv1.Job) (*batchv1.Job, error) {
	client, err := GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes client: %s", err)
	}

	return client.CreateJob(ctx, job)
}
