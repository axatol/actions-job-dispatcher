package server

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Response[T any] struct {
	Status  int    `json:"-"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	Error   error  `json:"-"`
}

func (r *Response[T]) Write(w http.ResponseWriter) {
	if r.Status < http.StatusOK {
		if r.Error != nil {
			r.Status = http.StatusInternalServerError
		} else {
			r.Status = http.StatusOK
		}
	}

	if r.Status == http.StatusOK && r.Message == "" {
		r.Message = "OK"
	}

	if r.Status >= http.StatusBadRequest && r.Message == "" {
		r.Message = "ERROR"
	}

	if r.Error != nil {
		log.Error().
			Err(r.Error).
			Int("status", r.Status).
			Str("message", r.Message).
			Msg("error handling request")
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

func responseWithMessage[T any](resp *Response[T], message ...string) *Response[T] {
	if len(message) == 1 {
		resp.Message = message[0]
	}

	return resp
}

func ResponseOK(message ...string) *Response[any] {
	resp := Response[any]{Status: http.StatusOK}
	return responseWithMessage(&resp, message...)
}

func ResponseErr(err error, message ...string) *Response[any] {
	return responseWithMessage(&Response[any]{Error: err}, message...)
}

func ResponseBadReq(err error, message ...string) *Response[any] {
	resp := Response[any]{
		Status: http.StatusBadRequest,
		Error:  err,
	}

	return responseWithMessage(&resp, message...)
}
