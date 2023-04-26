package server

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Response[T any] struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Error   error  `json:"-"`
}

func (r *Response[T]) Write(w http.ResponseWriter, logger ...zerolog.Logger) {
	log := log.Logger
	if len(logger) == 1 {
		log = logger[0]
	}

	if r.Status == http.StatusOK && r.Message == "" {
		r.Message = "OK"
	}

	if r.Status >= http.StatusBadRequest && r.Message == "" {
		r.Message = "ERROR"
	}

	if r.Error != nil {
		log.Error().Err(r.Error).Msg(r.Message)
	}

	raw, err := json.Marshal(r)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to marshal response data")
		r.Status = http.StatusInternalServerError
		raw = []byte(`{"message":"error handling request","data":null}`)
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(r.Status)
	if _, err := w.Write(raw); err != nil {
		log.Error().
			Int("status", r.Status).
			Str("message", r.Message).
			Err(err).
			Msg("failed to write response data")
		return
	}
}

func ResponseOK(message string) *Response[any] {
	return &Response[any]{
		Status:  http.StatusOK,
		Message: message,
	}
}

func ResponseErr(err error, message string) *Response[any] {
	return &Response[any]{
		Status:  http.StatusInternalServerError,
		Message: message,
		Error:   err,
	}
}

func ResponseBadReq(err error, message string) *Response[any] {
	return &Response[any]{
		Status:  http.StatusBadRequest,
		Message: message,
		Error:   err,
	}
}
