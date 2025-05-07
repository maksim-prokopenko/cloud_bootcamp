package roundrobin

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type SimpleCarouselllll interface {
	Set(url string, info any)
	Next() any
}

type roundRobin struct {
	newBackendsCarousel SimpleCarouselllll
	logger              *slog.Logger
}

type Config struct {
	Logger      *slog.Logger
	BackendUrls map[string]struct{}
}

func New(cfg Config) (http.Handler, error) {
	rr := &roundRobin{
		newBackendsCarousel: NewCarousel(),
	}

	var errs []error
	for backendUrl := range cfg.BackendUrls {
		backendUrlParsed, err := url.Parse(backendUrl)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(backendUrlParsed) // TODO inject logger and error handler in proxy

		rr.newBackendsCarousel.Set(backendUrl, proxy)
	}
	if len(cfg.BackendUrls) == 0 {
		errs = append(errs, errors.New("empty backends list"))
	}
	return rr, errors.Join(errs...)
}

func (rr *roundRobin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO handle error proxy when server on another port
	handler := rr.newBackendsCarousel.Next().(http.Handler)
	if handler == nil {
		rr.logger.Error("no one alive backend")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable) // TODO not json?
		return
	}
	handler.ServeHTTP(w, r)
}
