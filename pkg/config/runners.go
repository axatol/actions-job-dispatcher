package config

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type RunnerConfig struct {
	// github

	RunnerLabel string `yaml:"runner_label" json:"runner_label,omitempty"`

	// scheduler

	MaxReplicas int `yaml:"max_replicas" json:"max_replicas,omitempty"`
	// TODO? MinReplicas

	// kubernetes

	ServiceAccountName string `yaml:"service_account_name" json:"service_account_name,omitempty"`
	Image              string `yaml:"image" json:"image,omitempty"`
	Resources          struct {
		CPULimit      string `yaml:"cpu_limit"`
		MemoryLimit   string `yaml:"memory_limit"`
		CPURequest    string `yaml:"cpu_request"`
		MemoryRequest string `yaml:"memory_request"`
	} `yaml:"resources"`
}

func (c RunnerConfig) Validate() error {
	if c.RunnerLabel == "" {
		return fmt.Errorf("field required: RunnerLabel")
	}

	if _, err := resource.ParseQuantity(c.Resources.CPULimit); err != nil {
		return fmt.Errorf("invalid cpu limit: %s", err)
	}

	if _, err := resource.ParseQuantity(c.Resources.MemoryLimit); err != nil {
		return fmt.Errorf("invalid memory limit: %s", err)
	}

	if _, err := resource.ParseQuantity(c.Resources.CPURequest); err != nil {
		return fmt.Errorf("invalid cpu request: %s", err)
	}

	if _, err := resource.ParseQuantity(c.Resources.MemoryRequest); err != nil {
		return fmt.Errorf("invalid memory request: %s", err)
	}

	return nil
}

var (
	Runners []RunnerConfig
)
