package cache

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v51/github"
)

type WorkflowJobMeta struct {
	Organisation    string    `json:"organisation"`
	Repository      string    `json:"repository"`
	WorkflowID      int64     `json:"workflow_id"`
	WorkflowName    string    `json:"workflow_name"`
	WorkflowJobID   int64     `json:"workflow_job_id"`
	WorkflowJobName string    `json:"workflow_job_name"`
	WorkflowJobURL  string    `json:"workflow_job_url"`
	RunnerLabels    []string  `json:"runner_labels"`
	CreatedAt       time.Time `json:"queued_at"`
	StartedAt       time.Time `json:"started_at"`
}

func MetaFromStringMap(m map[string]string) WorkflowJobMeta {
	result := WorkflowJobMeta{}
	result.Organisation = m["organisation"]
	result.Repository = m["repository"]
	if workflowID, err := strconv.ParseInt(m["workflow-id"], 10, 64); err != nil {
		result.WorkflowID = workflowID
	}
	result.WorkflowName = m["workflow-name"]
	if workflowJobID, err := strconv.ParseInt(m["workflow-job-id"], 10, 64); err != nil {
		result.WorkflowJobID = workflowJobID
	}
	result.WorkflowJobName = m["workflow-job-name"]
	result.WorkflowJobURL = m["workflow-job-url"]
	result.RunnerLabels = strings.Split(m["runner-labels"], ",")
	return result
}

func (m WorkflowJobMeta) StringMap() map[string]string {
	result := map[string]string{}
	result["organisation"] = m.Organisation
	result["repository"] = m.Repository
	result["workflow-id"] = fmt.Sprint(m.WorkflowID)
	result["workflow-name"] = m.WorkflowName
	result["workflow-job-id"] = fmt.Sprint(m.WorkflowJobID)
	result["workflow-job-name"] = m.WorkflowJobName
	result["workflow-job-url"] = m.WorkflowJobURL
	result["runner-labels"] = strings.Join(m.RunnerLabels, ",")
	return result
}

func WorkflowJobMetaFromEvent(event *github.WorkflowJobEvent) *WorkflowJobMeta {
	m := WorkflowJobMeta{}
	m.Organisation = event.GetOrg().GetLogin()
	m.Repository = event.GetRepo().GetFullName()
	m.WorkflowID = event.GetWorkflowJob().GetRunID()
	m.WorkflowName = event.GetWorkflowJob().GetWorkflowName()
	m.WorkflowJobID = event.GetWorkflowJob().GetID()
	m.WorkflowJobName = event.GetWorkflowJob().GetName()
	m.WorkflowJobURL = event.GetWorkflowJob().GetHTMLURL()
	m.RunnerLabels = event.GetWorkflowJob().Labels
	return &m
}
