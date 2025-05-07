package roundrobin

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type roundRobin struct {
	backendsCarousel *backendsCarousel
	logger           *slog.Logger
}

type Config struct {
	Logger      *slog.Logger
	BackendUrls map[string]struct{}
}

func New(cfg Config) (http.Handler, error) {
	rr := &roundRobin{
		backendsCarousel: newBackendsCarousel(),
	}

	var errs []error
	for backendUrl := range cfg.BackendUrls {
		backendUrlParsed, err := url.Parse(backendUrl)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(backendUrlParsed) // TODO inject logger in proxy

		err = rr.backendsCarousel.addNew(StrUrl(backendUrl), proxy)
		if err != nil {
			errs = append(errs, err)
		}

	}
	return rr, errors.Join(errs...)
}

func (lb *roundRobin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := lb.backendsCarousel.nextHandler()
	if handler == nil {
		lb.logger.Error("no one alive backend")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable) // TODO not json?
		return
	}
	handler.ServeHTTP(w, r)
}
