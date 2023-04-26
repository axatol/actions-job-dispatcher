package config

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v51/github"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	Runners RunnerConfigList
)

type RunnerConfigList []RunnerConfig

func (rcl RunnerConfigList) Strs() []string {
	results := []string{}
	for _, runner := range rcl {
		results = append(results, runner.String())
	}

	return results
}

type RunnerConfig struct {
	// github

	Labels RunnerLabels `yaml:"labels" json:"labels,omitempty"`
	Scope  RunnerScope  `yaml:"scope"  json:"scope,omitempty"`

	// scheduler

	MaxReplicas int `yaml:"max_replicas" json:"max_replicas,omitempty"`
	// TODO? MinReplicas

	// kubernetes

	ServiceAccountName string          `yaml:"service_account_name" json:"service_account_name,omitempty"`
	Image              string          `yaml:"image"                json:"image,omitempty"`
	Resources          RunnerResources `yaml:"resources"            json:"resources,omitempty"`
}

func (c RunnerConfig) String() string {
	return fmt.Sprintf("%s:%s", c.Scope.String(), strings.Join(c.Labels, "+"))
}

func (c RunnerConfig) Slug() string {
	slug := c.String()
	slug = strings.ReplaceAll(slug, "/", "_") // repo delim
	slug = strings.ReplaceAll(slug, "+", "_") // scope delim
	slug = strings.ReplaceAll(slug, ":", "-") // label delim
	return slug
}

func (c RunnerConfig) Validate() error {
	if err := c.Labels.Validate(); err != nil {
		return fmt.Errorf("invalid labels: %s", err)
	}

	if err := c.Scope.Validate(); err != nil {
		return fmt.Errorf("invalid scope: %s", err)
	}

	if err := c.Resources.Validate(); err != nil {
		return fmt.Errorf("invalid resources: %s", err)
	}

	return nil
}

type RunnerLabels []string

func (rl RunnerLabels) String() string {
	return strings.Join(rl, ",")
}

func (rl RunnerLabels) Validate() error {
	if rl == nil || len(rl) < 1 {
		return fmt.Errorf("field required: labels")
	}

	return nil
}

func (rl RunnerLabels) Has(search string) bool {
	for _, label := range rl {
		if search == label {
			return true
		}
	}

	return false
}

type RunnerScope struct {
	Organisation string `yaml:"organisation" json:"organisation,omitempty"`
	Repository   string `yaml:"repository"   json:"repository,omitempty"`
}

func (rs RunnerScope) String() string {
	if rs.Organisation != "" {
		return rs.Organisation
	}

	return rs.Repository
}

func (rs RunnerScope) IsOrg() bool {
	return rs.Organisation != ""
}

func (rs RunnerScope) IsRepo() bool {
	return rs.Repository != ""
}

func (rs RunnerScope) GetRepo() (string, string, bool) {
	return strings.Cut(rs.Repository, "/")
}

func (rs RunnerScope) Validate() error {
	fields := []string{}

	if rs.Organisation != "" {
		fields = append(fields, rs.Organisation)
	}

	if rs.Repository != "" {
		fields = append(fields, rs.Repository)
	}

	if len(fields) != 1 {
		return fmt.Errorf(`must specify exactly one of "enterprise", "organisation", or "repository`)
	}

	return nil
}

func RunnerScopeFromWorkflowJobEvent(e *github.WorkflowJobEvent) (*RunnerScope, error) {
	scope := RunnerScope{}

	if org := e.GetOrg(); org != nil {
		scope.Organisation = *org.Login
	}

	if repo := e.GetRepo(); repo != nil {
		scope.Repository = *repo.FullName
	}

	if err := scope.Validate(); err != nil {
		return nil, err
	}

	return &scope, nil
}

type RunnerResources struct {
	CPULimit      string `yaml:"cpu_limit"      json:"cpu_limit"`
	MemoryLimit   string `yaml:"memory_limit"   json:"memory_limit"`
	CPURequest    string `yaml:"cpu_request"    json:"cpu_request"`
	MemoryRequest string `yaml:"memory_request" json:"memory_request"`
}

func (rr RunnerResources) Validate() error {
	if _, err := resource.ParseQuantity(rr.CPULimit); err != nil {
		return fmt.Errorf("invalid cpu limit: %s", err)
	}

	if _, err := resource.ParseQuantity(rr.MemoryLimit); err != nil {
		return fmt.Errorf("invalid memory limit: %s", err)
	}

	if _, err := resource.ParseQuantity(rr.CPURequest); err != nil {
		return fmt.Errorf("invalid cpu request: %s", err)
	}

	if _, err := resource.ParseQuantity(rr.MemoryRequest); err != nil {
		return fmt.Errorf("invalid memory request: %s", err)
	}

	return nil
}
