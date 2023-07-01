package server

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func middleware_Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Context().Value(middleware.RequestIDKey)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			event := log.Info()

			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			url := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

			if rec := recover(); rec != nil {
				event = log.Error().
					Interface("recovery", rec).
					Bytes("stack", debug.Stack())
			}

			event.
				Int("status", ww.Status()).
				Str("method", r.Method).
				Str("url", url).
				Str("proto", r.Proto).
				Dur("duration", time.Since(start)).
				Int("bytes_written", ww.BytesWritten()).
				Str("remote_addr", r.RemoteAddr).
				Str("origin", r.Header.Get("Origin")).
				Str("request_id", (requestID).(string)).
				Send()
		}()

		next.ServeHTTP(ww, r)
	})
}
