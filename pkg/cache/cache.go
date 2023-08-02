package cache

import (
	"time"

	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog/log"
)

var cache map[int64]WorkflowJobMeta

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

	cache[meta.WorkflowJobID] = meta
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

func CacheWorkflowJobEvent(event *github.WorkflowJobEvent) *WorkflowJobMeta {
	meta := Get(event.GetWorkflowJob().GetID())
	if meta == nil {
		meta = WorkflowJobMetaFromEvent(event)
	}

	status := event.GetWorkflowJob().GetStatus()
	switch status {
	case "queued":
		meta.CreatedAt = time.Now()
		Set(*meta)

	case "in_progress":
		meta.StartedAt = time.Now()
		Set(*meta)

	case "completed":
		Del(meta.WorkflowJobID)

	default:
		log.Warn().
			Str("status", status).
			Interface("meta", meta).
			Msg("unhandled workflow job status")
	}

	return meta
}
