package roundrobin

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/maximmihin/cb625/internal/balancer/roundrobin/carousel"
)

type RoundRobin struct {
	*carousel.Carousel
	logger *slog.Logger
}

type Config struct {
	BackendUrls []ServerConfig `json:"servers"`

	Logger *slog.Logger `json:"-"`
}

type ServerConfig struct {
	URL    string `json:"url"`
	Weight int    `json:"weight,omitempty"`
}

func New(cfg *Config) (*RoundRobin, error) {

	rr := new(RoundRobin)
	rr.Carousel = carousel.New()

	var errs []error
	for _, backendUrl := range cfg.BackendUrls {
		backendUrlParsed, err := url.Parse(backendUrl.URL)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(backendUrlParsed) // TODO inject logger and error handler in proxy

		rr.Carousel.SetWithWeight(backendUrl.URL, proxy, backendUrl.Weight)
	}
	if len(cfg.BackendUrls) == 0 {
		errs = append(errs, errors.New("empty backends list"))
	}
	return rr, errors.Join(errs...)
}

func (rr *RoundRobin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO handle error proxy when server on another port
	handler := rr.Carousel.Next().(http.Handler) // TODO handle case, when return nil (when all server died)
	if handler == nil {
		rr.logger.Error("no one alive backend")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable) // TODO not json?
		return
	}
	handler.ServeHTTP(w, r)
}
