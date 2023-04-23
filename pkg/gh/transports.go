package gh

import (
	"net/http"

	"github.com/gregjones/httpcache"
	"github.com/rs/zerolog/log"
)

const (
	headerRateLimitLimit     = "x-ratelimit-limit"
	headerRateLimitRemaining = "x-ratelimit-remaining"
	headerRateLimitUsed      = "x-ratelimit-used"
	headerRateLimitReset     = "x-ratelimit-reset"
)

type LoggingTransport struct{ Transport http.RoundTripper }

func (t LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.Transport.RoundTrip(req)
	t.log(req, res, err)
	return res, err
}

func (t LoggingTransport) log(req *http.Request, res *http.Response, err error) {
	event := log.Info()
	if err != nil {
		event = log.Error().Err(err)
	}

	if res != nil {
		event = event.
			Int("status_code", res.StatusCode).
			Bool("cache_hit", res.Header.Get(httpcache.XFromCache) == "1").
			Str("rate_limit_remaining", res.Header.Get(headerRateLimitRemaining))
	}

	event.
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Send()
}
