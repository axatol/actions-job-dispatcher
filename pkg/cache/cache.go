package cache

import (
	"time"

	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
)

var cache map[int64]WorkflowJobMeta

type WorkflowJobMeta struct {
	JobID     int64     `json:"job_id"`
	RunID     int64     `json:"run_id"`
	Name      string    `json:"name"`
	Labels    []string  `json:"runner_labels"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"queued_at"`
	StartedAt time.Time `json:"in_progress_at"`
}

func List() []WorkflowJobMeta {
	results := []WorkflowJobMeta{}
	for _, meta := range cache {
		results = append(results, meta)
	}

	return results
}

func Set(meta WorkflowJobMeta) WorkflowJobMeta {
	if cache == nil {
		cache = map[int64]WorkflowJobMeta{}
	}

	cache[meta.JobID] = meta
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

func CacheWorkflowJobEvent(event *github.WorkflowJobEvent) bool {
	meta := Get(event.WorkflowJob.GetID())
	if meta == nil {
		meta = &WorkflowJobMeta{
			JobID:  event.WorkflowJob.GetID(),
			RunID:  event.WorkflowJob.GetRunID(),
			Labels: event.WorkflowJob.Labels,
			URL:    event.WorkflowJob.GetHTMLURL(),
		}
	}

	status := event.WorkflowJob.GetStatus()
	switch status {
	case "in_progress":
		meta.StartedAt = time.Now()
		Set(*meta)
		return true

	case "queued":
		meta.CreatedAt = time.Now()
		Set(*meta)
		return true

	case "completed":
		Del(meta.JobID)
		return false

	default:
		log.Warn().
			Str("status", status).
			Interface("meta", meta).
			Msg("unhandled workflow job status")
		return false
	}
}
