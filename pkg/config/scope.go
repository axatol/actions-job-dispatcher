package config

import (
	"fmt"

	"github.com/google/go-github/v51/github"
)

type Scope struct {
	// TODO enterprise?

	IsOrg      bool   `yaml:"is_org"     json:"is_org,omitempty"`
	Owner      string `yaml:"owner"      json:"owner,omitempty"`
	Repository string `yaml:"repository" json:"repository,omitempty"`
}

func (rs Scope) String() string {
	if rs.IsOrg {
		return rs.Owner
	}

	return fmt.Sprintf("%s/%s", rs.Owner, rs.Repository)
}

func (rs Scope) Validate() error {
	if rs.Owner == "" {
		return fmt.Errorf("must specify owner")
	}

	if !rs.IsOrg && rs.Repository == "" {
		return fmt.Errorf("must specify repository if not an organisation scope")
	}

	return nil
}

func ScopeFromWorkflowJobEvent(e *github.WorkflowJobEvent) (*Scope, error) {
	scope := Scope{}
	scope.IsOrg = e.GetOrg() != nil
	scope.Owner = e.GetRepo().GetOwner().GetLogin()
	scope.Repository = e.GetRepo().GetName()

	if err := scope.Validate(); err != nil {
		return nil, err
	}

	return &scope, nil
}
