package cache

import (
	"time"

	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
)

var cache map[int64]WorkflowJobMeta

type WorkflowJobMeta struct {
	ID          int64     `json:"id"`
	RunID       int64     `json:"run_id"`
	Name        string    `json:"name"`
	RunnerLabel string    `json:"runner_label"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"queued_at"`
	StartedAt   time.Time `json:"in_progress_at"`
}

func List() []WorkflowJobMeta {
	results := []WorkflowJobMeta{}
	for _, meta := range cache {
		results = append(results, meta)
	}

	return results
}

func Set(meta WorkflowJobMeta) WorkflowJobMeta {
	cache[meta.ID] = meta
	return meta
}

func Get(id int64) *WorkflowJobMeta {
	if meta, ok := cache[id]; ok {
		return &meta
	}

	return nil
}

func Del(id int64) {
	delete(cache, id)
}

func HandleWorkflowJobEvent(event github.WorkflowJobEvent) {
	meta := Get(event.WorkflowJob.GetID())
	if meta == nil {
		meta = &WorkflowJobMeta{
			ID:          event.WorkflowJob.GetID(),
			RunID:       event.WorkflowJob.GetRunID(),
			RunnerLabel: event.WorkflowJob.Labels[1],
			URL:         event.WorkflowJob.GetHTMLURL(),
		}
	}

	status := event.WorkflowJob.GetStatus()
	switch status {
	case "in_progress":
		meta.StartedAt = time.Now()
		Set(*meta)
		return

	case "queued":
		meta.CreatedAt = time.Now()
		Set(*meta)
		return

	case "completed":
		Del(meta.ID)
		return

	default:
		log.Warn().
			Str("status", status).
			Interface("meta", meta).
			Msg("unhandled workflow job status")
	}
}
