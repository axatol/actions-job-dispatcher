package config

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

type RunnerConfigList []RunnerConfig

func (rcl RunnerConfigList) Validate() error {
	if len(rcl) < 1 {
		return fmt.Errorf("no runners configured")
	}

	for _, runner := range rcl {
		if err := runner.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (rcl RunnerConfigList) Strs() []string {
	results := []string{}
	for _, runner := range rcl {
		results = append(results, runner.String())
	}

	return results
}

type RunnerConfig struct {
	// github

	Labels Labels `yaml:"labels" json:"labels,omitempty"`
	Scope  Scope  `yaml:"scope"  json:"scope,omitempty"`

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

type Labels []string

func (rl Labels) String() string {
	return strings.Join(rl, ",")
}

func (rl Labels) Validate() error {
	if rl == nil || len(rl) < 1 {
		return fmt.Errorf("field required: labels")
	}

	return nil
}

func (rl Labels) Has(search string) bool {
	for _, label := range rl {
		if search == label {
			return true
		}
	}

	return false
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
